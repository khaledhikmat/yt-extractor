UPDATE videos 
SET 
    updated_at = NOW(),
    transcription_url = $1
WHERE id = $2

