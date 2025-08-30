CREATE TABLE posts
(
  id SERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  content TEXT NOT NULL,
  author VARCHAR(100) NOT NULL,
  created_at TIMESTAMP
  WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
  WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

  -- Create index on created_at for better performance on sorting
  CREATE INDEX idx_posts_created_at ON posts(created_at DESC);

  -- Create index on author for filtering
  CREATE INDEX idx_posts_author ON posts(author);
