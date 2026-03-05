-- Story 2.2 rollback: remove users.status field
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'users'
    ) THEN
        IF EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'status'
        ) THEN
            ALTER TABLE public.users DROP COLUMN status;
        END IF;
    END IF;
END $$;
