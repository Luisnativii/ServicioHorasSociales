'use strict';

const types = { feature: 'feat', fix: 'fix', chore: 'chore', docs: 'docs', refactor: 'refactor', test: 'test', ci: 'ci' };
const tagPattern = /^V(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/;

function parseTag(tag) {
  const match = tagPattern.exec(tag);
  if (!match) throw new Error(`Tag SemVer inválido: ${tag}`);
  return match.slice(1).map(Number);
}
function compareTags(a, b) {
  const aa = parseTag(a), bb = parseTag(b);
  return aa[0] - bb[0] || aa[1] - bb[1] || aa[2] - bb[2];
}
function bump(tag, level = 'patch') {
  const [major, minor, patch] = parseTag(tag);
  if (level === 'major') return `V${major + 1}.0.0`;
  if (level === 'minor') return `V${major}.${minor + 1}.0`;
  if (level === 'patch') return `V${major}.${minor}.${patch + 1}`;
  throw new Error('Incremento de versión inválido');
}
function labelsOf(item) { return (item.labels || []).map(label => typeof label === 'string' ? label : label.name); }
function issueType(issue) {
  const selected = labelsOf(issue).filter(label => label.startsWith('type:'));
  if (selected.length !== 1 || !Object.hasOwn(types, selected[0].slice(5))) throw new Error('Selecciona exactamente una etiqueta type: válida');
  return selected[0].slice(5);
}
function branchForIssue(issue) {
  if (issue.pull_request || issue.state !== 'open') throw new Error('Se requiere un issue abierto, no un PR');
  const type = issueType(issue);
  const slug = issue.title.normalize('NFKD').replace(/[\u0300-\u036f]/g, '').toLowerCase()
    .replace(/^(feat|fix|chore|docs|refactor|test|ci)(\([^)]*\))?:\s*/, '')
    .replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '').slice(0, 60).replace(/-$/g, '') || 'tarea';
  return `${type}/${issue.number}-${slug}`;
}
function versionLevel(pr) {
  const versions = labelsOf(pr).filter(label => label.startsWith('version:'));
  if (versions.length !== 1 || !/^version:(patch|minor|major)$/.test(versions[0])) throw new Error('Selecciona una sola etiqueta version:patch, version:minor o version:major');
  return versions[0].slice(8);
}
function validatePullRequest(pr, issue, merged = false) {
  if (pr.base.ref !== 'develop') throw new Error('main está reservada; el destino del PR debe ser develop');
  const match = /^(feature|fix|chore|docs|refactor|test|ci)\/([1-9]\d*)-([a-z0-9]+(?:-[a-z0-9]+)*)$/.exec(pr.head.ref);
  if (!match || match[3].length > 60) throw new Error('Rama inválida; usa tipo/NUMERO-descripcion (máximo 60 caracteres de descripción)');
  const number = Number(match[2]);
  if (!issue || issue.number !== number || issue.pull_request || (!merged && issue.state !== 'open')) throw new Error('La rama debe referenciar un issue abierto del mismo repositorio');
  if (issueType(issue) !== match[1]) throw new Error('El tipo de rama no coincide con la etiqueta type: del issue');
  const title = /^(feat|fix|chore|docs|refactor|test|ci)(\([a-z0-9-]+\))?(!)?: .+ \(#([1-9]\d*)\)$/.exec(pr.title);
  if (!title || title[1] !== types[match[1]] || Number(title[4]) !== number) throw new Error('Título inválido; ejemplo: feat(api): describir el cambio (#12)');
  if (!(new RegExp(`^Closes #${number}\\s*$`, 'm')).test(pr.body || '')) throw new Error(`Agrega una línea Closes #${number} en la descripción`);
  const level = versionLevel(pr);
  if (title[3] && level !== 'major') throw new Error('Un cambio incompatible (!) requiere version:major');
  return { number, level };
}
async function tagCommit(github, repo, ref) {
  if (ref.object.type === 'tag') return (await github.rest.git.getTag({ ...repo, tag_sha: ref.object.sha })).data.object.sha;
  return ref.object.sha;
}
async function allTags(github, repo) {
  const refs = await github.paginate(github.rest.git.listMatchingRefs, { ...repo, ref: 'tags/V' });
  return refs.filter(ref => tagPattern.test(ref.ref.replace('refs/tags/', ''))).sort((a, b) => compareTags(b.ref.replace('refs/tags/', ''), a.ref.replace('refs/tags/', '')));
}
async function planVersion({ github, context }) {
  const repo = context.repo;
  const tags = await allTags(github, repo);
  if (!tags.length) throw new Error('Falta el tag inicial V0.0.1; no se inventará una versión de partida');
  const previous = tags[0].ref.replace('refs/tags/', '');
  if (context.eventName === 'pull_request') {
    const pr = (await github.rest.pulls.get({ ...repo, pull_number: context.payload.pull_request.number })).data;
    const match = /^(?:feature|fix|chore|docs|refactor|test|ci)\/([1-9]\d*)-/.exec(pr.head.ref);
    if (!match) throw new Error('La rama debe comenzar con tipo/NUMERO-');
    const issue = (await github.rest.issues.get({ ...repo, issue_number: Number(match[1]) })).data;
    const { level } = validatePullRequest(pr, issue);
    const tag = bump(previous, level);
    return { tag, version: tag.slice(1), release: false, previous };
  }
  if (context.ref !== 'refs/heads/develop') throw new Error('El versionado solo se ejecuta desde develop');
  for (const tag of tags) {
    if (await tagCommit(github, repo, tag) === context.sha) {
      const name = tag.ref.replace('refs/tags/', '');
      return { tag: name, version: name.slice(1), release: false, previous: name };
    }
  }
  const associated = await github.paginate(github.rest.repos.listPullRequestsAssociatedWithCommit, { ...repo, commit_sha: context.sha });
  let pr;
  for (const candidate of associated.filter(item => item.base.ref === 'develop')) {
    const full = (await github.rest.pulls.get({ ...repo, pull_number: candidate.number })).data;
    if (full.merged && full.merge_commit_sha === context.sha) { pr = full; break; }
  }
  if (!pr) throw new Error('develop acepta integraciones por PR; este commit no proviene de un PR integrado');
  const number = Number(pr.head.ref.split('/')[1]?.split('-')[0]);
  const issue = (await github.rest.issues.get({ ...repo, issue_number: number })).data;
  const { level } = validatePullRequest(pr, issue, true);
  const comparison = (await github.rest.repos.compareCommits({ ...repo, base: previous, head: context.sha })).data;
  if (!['ahead', 'identical'].includes(comparison.status)) throw new Error('La versión anterior no pertenece al historial de develop');
  const tag = bump(previous, level);
  return { tag, version: tag.slice(1), release: true, previous };
}
async function publishVersion({ github, context, core, plan }) {
  if (!plan.release || context.ref !== 'refs/heads/develop' || context.eventName === 'pull_request') throw new Error('Publicación de versión no autorizada en este evento');
  const head = (await github.rest.git.getRef({ ...context.repo, ref: 'heads/develop' })).data.object.sha;
  if (head !== context.sha) {
    core.info('La ejecución quedó atrás de develop; se etiquetará la siguiente integración vigente.');
    return { skipped: true };
  }
  const tags = await allTags(github, context.repo);
  const existing = tags.find(ref => ref.ref === `refs/tags/${plan.tag}`);
  if (existing) {
    if (await tagCommit(github, context.repo, existing) !== context.sha) throw new Error('El tag ya existe en otro commit; no se sobrescribirá');
    return { tag: plan.tag, existing: true };
  }
  if (tags[0]?.ref !== `refs/tags/${plan.previous}`) throw new Error('Cambió la última versión durante CI; vuelve a ejecutar la pipeline');
  const annotated = await github.rest.git.createTag({ ...context.repo, tag: plan.tag, message: `${plan.tag}: integración validada en develop`, object: context.sha, type: 'commit' });
  await github.rest.git.createRef({ ...context.repo, ref: `refs/tags/${plan.tag}`, sha: annotated.data.sha });
  await core.summary.addHeading(`Versión ${plan.tag}`).addRaw(`Commit validado: ${context.sha}`).write();
  return { tag: plan.tag };
}
async function manageIssue({ github, context, core }) {
  const repo = context.repo;
  const issue = (await github.rest.issues.get({ ...repo, issue_number: context.payload.issue.number })).data;
  const remove = async name => {
    try { await github.rest.issues.removeLabel({ ...repo, issue_number: issue.number, name }); }
    catch (error) { if (error.status !== 404) throw error; }
  };
  if (context.payload.action === 'closed') {
    for (const name of ['status:backlog', 'status:ready', 'status:in-progress']) await remove(name);
    await github.rest.issues.addLabels({ ...repo, issue_number: issue.number, labels: ['status:done'] });
    return;
  }
  if (context.payload.action === 'reopened') {
    await remove('status:done');
    await github.rest.issues.addLabels({ ...repo, issue_number: issue.number, labels: ['status:backlog'] });
    return;
  }
  if (context.payload.label?.name !== 'status:ready' || issue.state !== 'open') return;
  const permission = (await github.rest.repos.getCollaboratorPermissionLevel({ ...repo, username: context.actor })).data.permission;
  if (!['admin', 'maintain', 'write'].includes(permission)) throw new Error('Solo colaboradores con escritura pueden iniciar una tarea');
  const desired = branchForIssue(issue);
  const type = issueType(issue);
  const existing = await github.paginate(github.rest.git.listMatchingRefs, { ...repo, ref: `heads/${type}/${issue.number}-` });
  const branch = existing.length ? existing[0].ref.replace('refs/heads/', '') : desired;
  if (!existing.length) {
    const base = (await github.rest.git.getRef({ ...repo, ref: 'heads/develop' })).data.object.sha;
    await github.rest.git.createRef({ ...repo, ref: `refs/heads/${branch}`, sha: base });
  }
  await remove('status:backlog');
  await remove('status:ready');
  await github.rest.issues.addLabels({ ...repo, issue_number: issue.number, labels: ['status:in-progress'] });
  const marker = '<!-- hs:issue-branch -->';
  const body = `${marker}\nRama lista: [${branch}](https://github.com/${repo.owner}/${repo.repo}/tree/${branch}).\n\nParte de \`develop\`. El PR debe apuntar a \`develop\`, incluir \`Closes #${issue.number}\` y una etiqueta \`version:patch\`, \`version:minor\` o \`version:major\`.\n\nPara usarla localmente:\n\n\`\`\`bash\ngit fetch origin --prune --tags\ngit switch --track origin/${branch}\n\`\`\``;
  const comments = await github.paginate(github.rest.issues.listComments, { ...repo, issue_number: issue.number });
  const comment = comments.find(item => item.user?.login === 'github-actions[bot]' && item.body?.includes(marker));
  if (comment) await github.rest.issues.updateComment({ ...repo, comment_id: comment.id, body });
  else await github.rest.issues.createComment({ ...repo, issue_number: issue.number, body });
  core.info(`Rama de trabajo: ${branch}`);
}
module.exports = { parseTag, compareTags, bump, branchForIssue, versionLevel, validatePullRequest, planVersion, publishVersion, manageIssue };
