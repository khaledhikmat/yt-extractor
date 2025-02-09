UPDATE jobs 
SET 
    state = $1, 
    videos = $2, 
    errors = $3, 
    completed_at = $4
WHERE id = $5
