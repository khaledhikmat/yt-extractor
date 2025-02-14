ALTER TABLE videos
ADD COLUMN audioed_at TIMESTAMP,
ADD COLUMN transcribed_at TIMESTAMP;

UPDATE videos 
SET audioed_at = processed_at, transcribed_at = processed_at 
WHERE processed_at IS NOT NULL;

ALTER TABLE videos 
DROP COLUMN processed_at;