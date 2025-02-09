UPDATE videos 
SET 
    updated_at = NOW(),
    extracted_at = NOW(),
    extraction_url = $1 
WHERE id = $2

