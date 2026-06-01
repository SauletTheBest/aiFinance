DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_user'
    ) THEN
        ALTER TABLE transactions
            DROP CONSTRAINT fk_user;
    END IF;
END $$;

ALTER TABLE users
    ALTER COLUMN id DROP DEFAULT;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'transactions'
          AND column_name = 'user_id'
          AND udt_name = 'uuid'
    ) THEN
        ALTER TABLE transactions
            ALTER COLUMN user_id TYPE TEXT USING user_id::text;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'users'
          AND column_name = 'id'
          AND udt_name = 'uuid'
    ) THEN
        ALTER TABLE users
            ALTER COLUMN id TYPE TEXT USING id::text;
    END IF;
END $$;
