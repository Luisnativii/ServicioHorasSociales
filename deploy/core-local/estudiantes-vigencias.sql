-- Contrato de ensayo #55. En la base compartida lo aplica el operador del core.
BEGIN;
SELECT pg_advisory_xact_lock(55);
CREATE TABLE IF NOT EXISTS public.student_career_periods (
    student_id uuid NOT NULL REFERENCES public.students(id) ON DELETE CASCADE,
    career_id uuid NOT NULL REFERENCES public.careers(id),
    faculty_id uuid NOT NULL REFERENCES public.faculties(id),
    valid_from timestamptz,
    valid_until timestamptz,
    actor text,
    reason text,
    CHECK (valid_from IS NULL OR valid_until IS NULL OR valid_from < valid_until)
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_student_career_period_open
    ON public.student_career_periods (student_id) WHERE valid_until IS NULL;
CREATE INDEX IF NOT EXISTS ix_student_career_period_history
    ON public.student_career_periods (student_id, valid_from);

CREATE OR REPLACE FUNCTION public.fn_student_career_periods()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $$
DECLARE
    effective_at timestamptz := clock_timestamp();
    previous_start timestamptz;
BEGIN
    IF TG_OP = 'UPDATE' THEN
        IF NEW.career_id = OLD.career_id THEN
            RETURN NEW;
        END IF;
        SELECT valid_from INTO previous_start
        FROM public.student_career_periods
        WHERE student_id = OLD.id AND valid_until IS NULL;
        IF FOUND THEN
            IF previous_start IS NOT NULL THEN
                effective_at := greatest(effective_at, previous_start + interval '1 microsecond');
            END IF;
            UPDATE public.student_career_periods SET valid_until = effective_at
            WHERE student_id = OLD.id AND valid_until IS NULL;
        ELSE
            INSERT INTO public.student_career_periods (student_id, career_id, faculty_id, valid_until)
            SELECT OLD.id, c.id, c.faculty_id, effective_at FROM public.careers c WHERE c.id = OLD.career_id;
        END IF;
    END IF;
    INSERT INTO public.student_career_periods
        (student_id, career_id, faculty_id, valid_from, actor, reason)
    SELECT NEW.id, c.id, c.faculty_id, effective_at,
           nullif(current_setting('app.actor', true), ''),
           nullif(current_setting('app.career_change_reason', true), '')
    FROM public.careers c WHERE c.id = NEW.career_id;
    RETURN NEW;
END;
$$;

REVOKE ALL ON FUNCTION public.fn_student_career_periods() FROM PUBLIC;
DROP TRIGGER IF EXISTS trg_student_career_periods ON public.students;
CREATE TRIGGER trg_student_career_periods
AFTER INSERT OR UPDATE OF career_id ON public.students
FOR EACH ROW EXECUTE FUNCTION public.fn_student_career_periods();
GRANT SELECT (id, carnet, nombres, apellidos, nombre_completo, correo_institucional, career_id, activo),
      INSERT (carnet, nombres, apellidos, correo_institucional, career_id),
      UPDATE (carnet, nombres, apellidos, correo_institucional, career_id)
    ON public.students TO :"app_role";
GRANT SELECT ON public.student_career_periods TO :"app_role";
COMMIT;
