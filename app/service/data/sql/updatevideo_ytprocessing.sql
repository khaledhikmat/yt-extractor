UPDATE videos 
SET 
    updated_at = NOW(),
    processed_at = NOW()
WHERE id = $1

