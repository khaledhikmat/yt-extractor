SELECT count(*) from videos where extraction_url is null;
SELECT count(*) from videos;
SELECT count(*) from videos where externalized_at is not null;
SELECT count(*) from videos where externalized_at is null;
SELECT count(*) from videos where processed_at is not null;
SELECT count(*) from videos where audio_url is not null;
SELECT count(*) from videos where published_at >= '2025-01-01 00:00:00';
SELECT count(*) from videos where published_at >= '2025-01-01 00:00:00' AND audio_url is not null;
SELECT * from videos where published_at >= '2025-01-01 00:00:00' AND audio_url is null;
select count(*) from jobs where state = 'completed';
select * from videos order by published_at desc limit 50;
select * from videos where video_url IN ('https://www.youtube.com/watch?v=aUSJG8AI05g', 'https://www.youtube.com/watch?v=J2f8LUZFcD8');
select * from videos where extraction_url = 'https://www.isitdownrightnow.com';
SELECT * from videos where published_at >= '2025-01-01 00:00:00' ORDER BY published_at DESC;
SELECT title, extraction_url, extracted_at, audio_url, audioed_at, transcription_url, transcribed_at from videos WHERE id = 42;
-- update videos SET extraction_url = 'https://www.isitdownrightnow.com' WHERE id = 1219 
-- update videos SET processed_at = null, externalized_at = null;
select * from jobs where state = 'running';
--delete from jobs where state = 'running';
select * from jobs where id = 590;
select * from jobs where state = 'completed' ORDER BY started_at DESC LIMIT 10;
select * from jobs ORDER BY started_at DESC LIMIT 10;
select * from errors;
-- delete from errors;

