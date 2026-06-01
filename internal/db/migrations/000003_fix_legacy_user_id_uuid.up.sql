DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'users'
          AND column_name = 'id'
          AND udt_name = 'text'
    ) THEN
        ALTER TABLE users
            ALTER COLUMN id TYPE UUID USING id::uuid;
    END IF;
END $$;

ALTER TABLE users
    ALTER COLUMN id SET DEFAULT uuid_generate_v4();

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'transactions'
          AND column_name = 'user_id'
          AND udt_name = 'text'
    ) THEN
        ALTER TABLE transactions
            ALTER COLUMN user_id TYPE UUID USING user_id::uuid;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_user'
    ) THEN
        ALTER TABLE transactions
            ADD CONSTRAINT fk_user
            FOREIGN KEY (user_id)
            REFERENCES users(id)
            ON DELETE CASCADE;
    END IF;
END $$;
