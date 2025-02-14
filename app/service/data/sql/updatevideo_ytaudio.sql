UPDATE videos 
SET 
    updated_at = NOW(),
    audioed_at = NOW(),
    audio_url = $1
WHERE id = $2

