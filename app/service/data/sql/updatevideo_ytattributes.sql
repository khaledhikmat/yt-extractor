UPDATE videos 
SET 
    updated_at = NOW(),
    views = $1, 
    comments = $2, 
    likes = $3 
WHERE id = $4
