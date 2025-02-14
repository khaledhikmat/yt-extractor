The `yt-extractor` project extracts yt videos given a channel ID, stores video metadata, uploads video, audio and transcription files to AWS S3 storage. 

The project provides a user interface on [Notion](https://notion.com).

## Macro Architecture

- Backend
    - Golang
    - Railway
- Database
    - Postgres
    - Neon
- Automations
    - Make.com
    - CloudConvert
    - OpenAI
- Frontend
    - Notion
    - Google Sheet

## Tools

The following tools are used in this project:

| Tool            | Description                       | Fee |
|-----------------|-----------------------------------|----------|
| [Make.com](https://us2.make.com) | Automation Platform | $9 Monthly for 10,000 ops |
| [Railway](https://railway.com/) | App Deployment Platform: App + Postgres Database | $5 Monthly for 8 GB/8 vCPU |
| [OpenAI](https://platform.openai.com/) | AI Platform | Pay-as-you go - Auto-Recharge enabled to maintain $20 balance |
| [CloudConvert](https://cloudconvert.com) | File Conversion Platform | Pay-as-you-go - Auto-Recharge enabled $9 for 500 conversions |
| [Notion](https://notion.com) | Wiki, Databases, Sites, etc Platform | $10 Monthly |
| [Google Sheets](https://docs.google.com/spreadsheets) | Spreadsheet | Free Tier |

## Deployment

Currently the deployment is manual. But the following are some improvements:

### AWS

- Terraform is not used. The only AWS resource is an S3 bucket.
- S3 bucket `yt-extractor` is created manually from the console and it is publicly accessible. 

### Neon

- Make use of branches
- Install CLI
- Automate using API
- Backup the database:

```bash
pg_dump -h ep-crimson-feather-a5n37o2g-pooler.us-east-2.aws.neon.tech -U neondb_owner -d neondb -F c -f ./dba/dumps/neon_backup_$(date +"%Y-%m-%d").dump
```

```bash
pg_dump -h ep-crimson-feather-a5n37o2g-pooler.us-east-2.aws.neon.tech -U neondb_owner -d neondb -F p -f ./dba/dumps/neon_backup_$(date +"%Y-%m-%d").sql
```

### Railway

- Install CLI
- Automate Deployment using API
- Export the database:

```bash
pg_dump -h monorail.proxy.rlwy.net -p 11397 -U postgres -d railway -F c -f ./dba/dumps/railway_backup_$(date +"%Y-%m-%d").dump
```

- Import the database:

```bash
pg_restore --host=monorail.proxy.rlwy.net --port=11397 --username=postgres --dbname=railway --format=c ./dba/dumps/backup_2025-02-13.dump
```

```bash
psql --host=monorail.proxy.rlwy.net --port=11397 --username=postgres --dbname=railway -f ./dba/dumps/backup_2025-02-13.sql
```

### Make.com

- Stop Automations via API
- Start Automations via API

## Issues

- ~~Need a new Github repo.~~
- Dockefile must be optimized.
- Debug statement appear not strcutured in Railway log output.
- ~~Need a tool to prepare for deployment:~~
    - ~~Remove all bucket entries.~~
    - ~~Truncate database tables.~~
    - ~~Delete Google Sheet rows.~~
    - ~~Delete Notion Rows.~~
- ~~Observability? We have errors table in addition to Railway.~~
- ~~Make.com automations require variables to store channel ID and API KEY. It turned out there is something called scenario input that is useful in this case.~~
- ~~Add notes on the automations.~~
- ~~Add Reset Factory API?~~
- ~~Audio Split to generate transcription.~~
- ~~Video Summary in Arabic and English is no longer needed.~~
- ~~Video ID `E8yRq75_yBo` is being converted by the Cloud Convert to `E8yRq75yBo`!!!! This is solved by using CloudConvert file name instead of the video ID.~~
- ~~Make.com automations must use both Google Sheets and Notion.~~
- ~~Move transcription to the backend. [Make.com](https://us2.make.com) is a really nice platform but it can be cost prohibitive especially when the data transfer (i.e. egress) becomes big. Since transcription of MP4 files involve sending/downloading large files to/from CloudConvert and OpenAI, it is probably easier to run transcription locally on the backend.~~
- ~~Considered AWS S3 Auto-Transcription but dismissed it due to cost.~~
- ~~OpenAI does not aupport async calling where results are rendered via webhook.~~
- Not sure how to mitigate the job running risk
- Consider using webhook for CloudConvert PROD and polling for DEV. 
- Automation risk where if insert/update fails to external databases. We may insert another time.  

## Automations

These automations require a Youtube channel ID to operarte on and an API Key:

| Automation      | Description                       | Interval | Size |
|-----------------|-----------------------------------|----------|------|
| Pull            | Request yt videos be pulled from Youtube using API  | Daily at 6:00 AM | 100 |
| Extract         | Request unextracted yt videos be extracted into S3   | Daily at 7:00 AM | 10 |
| Re-attempt Extract | Request errored extractions be re-attempted   | Daily at 8:00 AM | 10 |
| Audio | Request yt videos be audioed   | Daily at 9:00 AM | 10 |
| Re-attempt Audio | Request errored audios be re-attempted   | Daily at 10:00 AM | 10 |
| Transcribe | Request yt videos be transcribed   | Daily at 11:00 AM | 10 |
| Re-attempt Transcribe | Request errored transcriptions be re-attempted   | Daily at 12:00 PM | 10 |
| Externalization | Export extracted videos to external sheets (Google and Notion)   | Daily at 1:00 PM  | 100 |
| Updation | Updates any updated records in the last 24 hrs to set the latest video metrics: comments, views and likes in addition to the audio, transcription and extraction URLs  | Daily at 2:00 PM  | 100 |

### Pipeline

Pull -> Extract -> Externalize -> Audio -> Transcribe

### Make.com

- There is no way to insert into [Google Sheets](https://docs.google.com/spreadsheets) and [Notion](https://notion.com) in parallel steps (using a Router for example) and then update the backend database as in the `yt-extractor-externalization` automation, for example. So I insert them in series and then call the backend:

```
Insert into Google Sheet => Insert into Notion => Update the backend
```

It would have been ideal to use a flow control like `merge` that would wait on the parallel tasks to execute before processding to the next step. But this is not available in Make.com. 

### Notion

- I imported [Google Sheets](https://docs.google.com/spreadsheets) (after I downloaded it as CSV) into Notion to create the initial videos database. 
- [Notion](https://notion.com) requires that one field be called `title`. Since we have a `title` in our database, I changed the field name to `summary`.
- I think it is probably best to build the database manually instead of importing. The import made the `video_id` field to be `title` and this cannot be changed. I prefer than it be `id`.  
- In [Notion](https://notion.com), I also adjusted the column types especially date and time.

## Extraction

The best way to view the downloaded MP4 files is a [VLC Player for MacOS](https://www.videolan.org/vlc/download-macosx.html) or [VLC Player for Windows](https://www.videolan.org/vlc/download-windows.html).

Youtube MP4 files can be either encoded using `AV1` or `H264` codec. `H264` is widely supported across most devices and players. MacOS Quicktime, for example, does not natively support `AV1` formatting.  

The following are some examples of how the streams look like in MP4 files when examined by `ffbrobe` which is part of `ffmpeg`:

- File encoded using H264 codec:

*Input*

```bash
ffprobe -v error -show_streams -select_streams v -of json "https://yt-extractor.s3.us-east-2.amazonaws.com/UCP-PfkMcOKriSxFMH7pTxfA/0GL5B003zYM.mp4"
```

*Output*

```json
{
    "streams": [
        {
            "index": 0,
            "codec_name": "h264",
            "codec_long_name": "H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10",
            "profile": "High",
            "codec_type": "video",
            "codec_tag_string": "avc1",
            "codec_tag": "0x31637661",
            "width": 1920,
            "height": 1080,
            "coded_width": 1920,
            "coded_height": 1080,
            "closed_captions": 0,
            "film_grain": 0,
            "has_b_frames": 1,
            "sample_aspect_ratio": "1:1",
            "display_aspect_ratio": "16:9",
            "pix_fmt": "yuv420p",
            "level": 40,
            "color_range": "tv",
            "color_space": "bt709",
            "color_transfer": "bt709",
            "color_primaries": "bt709",
            "chroma_location": "left",
            "field_order": "progressive",
            "refs": 1,
            "is_avc": "true",
            "nal_length_size": "4",
            "id": "0x1",
            "r_frame_rate": "25/1",
            "avg_frame_rate": "25/1",
            "time_base": "1/12800",
            "start_pts": 0,
            "start_time": "0.000000",
            "duration_ts": 19580928,
            "duration": "1529.760000",
            "bit_rate": "1008988",
            "bits_per_raw_sample": "8",
            "nb_frames": "38244",
            "extradata_size": 45,
            "disposition": {
                "default": 1,
                "dub": 0,
                "original": 0,
                "comment": 0,
                "lyrics": 0,
                "karaoke": 0,
                "forced": 0,
                "hearing_impaired": 0,
                "visual_impaired": 0,
                "clean_effects": 0,
                "attached_pic": 0,
                "timed_thumbnails": 0,
                "non_diegetic": 0,
                "captions": 0,
                "descriptions": 0,
                "metadata": 0,
                "dependent": 0,
                "still_image": 0,
                "multilayer": 0
            },
            "tags": {
                "language": "und",
                "handler_name": "ISO Media file produced by Google Inc.",
                "vendor_id": "[0][0][0][0]"
            }
        }
    ]
}
```

- File encoded using AV1 codec:

*Input*

```bash
ffprobe -v error -show_streams -select_streams v -of json "https://yt-extractor.s3.us-east-2.amazonaws.com/UCP-PfkMcOKriSxFMH7pTxfA/0hiQBvg5OmY.mp4"
```

*Output*

```json
{
    "streams": [
        {
            "index": 0,
            "codec_name": "av1",
            "codec_long_name": "Alliance for Open Media AV1",
            "profile": "Main",
            "codec_type": "video",
            "codec_tag_string": "av01",
            "codec_tag": "0x31307661",
            "width": 1280,
            "height": 720,
            "coded_width": 1280,
            "coded_height": 720,
            "closed_captions": 0,
            "film_grain": 0,
            "has_b_frames": 0,
            "pix_fmt": "yuv420p",
            "level": 5,
            "color_range": "tv",
            "color_space": "bt709",
            "color_transfer": "bt709",
            "color_primaries": "bt709",
            "refs": 1,
            "id": "0x1",
            "r_frame_rate": "25/1",
            "avg_frame_rate": "25/1",
            "time_base": "1/12800",
            "start_pts": 0,
            "start_time": "0.000000",
            "duration_ts": 40790528,
            "duration": "3186.760000",
            "bit_rate": "501987",
            "nb_frames": "79669",
            "extradata_size": 20,
            "disposition": {
                "default": 1,
                "dub": 0,
                "original": 0,
                "comment": 0,
                "lyrics": 0,
                "karaoke": 0,
                "forced": 0,
                "hearing_impaired": 0,
                "visual_impaired": 0,
                "clean_effects": 0,
                "attached_pic": 0,
                "timed_thumbnails": 0,
                "non_diegetic": 0,
                "captions": 0,
                "descriptions": 0,
                "metadata": 0,
                "dependent": 0,
                "still_image": 0,
                "multilayer": 0
            },
            "tags": {
                "language": "und",
                "handler_name": "ISO Media file produced by Google Inc.",
                "vendor_id": "[0][0][0][0]"
            }
        }
    ]
}
```

Here is what you can do to convert from AV1 to H264:

```bash
# convert from AV1 to H264:
ffmpeg -i "https://yt-extractor.s3.us-east-2.amazonaws.com/UCP-PfkMcOKriSxFMH7pTxfA/14NYvRyAe3Y.mp4" -c:v libx264 -preset fast -crf 23 -c:a copy "output_file.mp4"
```

## Risks

- Google YT API Key expiration in TEST mode. 
- Extractor tool constant updates may require periodic deployments.