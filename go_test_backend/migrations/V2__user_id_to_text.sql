DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'users'
      AND column_name = 'user_id'
      AND data_type <> 'text'
  ) THEN
    ALTER TABLE users
      ALTER COLUMN user_id TYPE TEXT
      USING user_id::text;
  END IF;
END $$;

