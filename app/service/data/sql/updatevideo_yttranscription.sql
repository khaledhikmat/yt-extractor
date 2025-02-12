UPDATE videos 
SET 
    updated_at = NOW(),
    processed_at = NOW(),
    audio_url = $1,
    transcription_url = $2
WHERE id = $3

