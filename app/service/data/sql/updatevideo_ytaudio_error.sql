UPDATE videos 
SET 
    updated_at = NOW(),
    audio_url = $1
WHERE id = $2

