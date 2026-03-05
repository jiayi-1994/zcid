-- Story 2.2: add status field for user account lifecycle management
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'users'
    ) THEN
        IF NOT EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'status'
        ) THEN
            ALTER TABLE public.users
                ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active';
        END IF;

        UPDATE public.users
        SET status = 'active'
        WHERE status IS NULL;
    END IF;
END $$;
