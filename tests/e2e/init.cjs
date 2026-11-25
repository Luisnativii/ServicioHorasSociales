const fs = require('node:fs');
const { randomBytes } = require('node:crypto');

for (const [name, content] of [
  ['jwt', randomBytes(32)],
  ['a.password', '8' + randomBytes(20).toString('hex')],
  ['b.password', '9' + randomBytes(20).toString('hex')],
]) {
  const path = '/e2e-secrets/' + name;
  try {
    fs.writeFileSync(path, content, { flag: 'wx', mode: 0o600 });
  } catch (error) {
    if (error.code !== 'EEXIST') throw error;
  }
  fs.chownSync(path, 65532, 65532);
}
