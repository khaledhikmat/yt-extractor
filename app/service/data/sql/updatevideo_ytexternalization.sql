UPDATE videos 
SET 
    updated_at = NOW(),
    externalized_at = NOW()
WHERE id = $1

