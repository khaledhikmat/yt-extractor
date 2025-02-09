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
| [Railway](https://railway.com/) | App Deployment Platform | $5 Monthly for 8 GB/8 vCPU |
| [Neon](https://console.neon.tech/app/projects) | Postgres Platform | Free Tier |
| [OpenAI](https://platform.openai.com/) | AI Platform | Pay-as-you go - Auto-Recharge enabled to maintain $20 balance |
| [CloudConvert](https://cloudconvert.com) | File Conversion Platform | Pay-as-you-go - Auto-Recharge enabled $9 for 500 conversions |
| [Notion](https://notion.com) | Wiki, Databases, Sites, etc Platform | $10 Monthly |
| [Google Sheets](https://docs.google.com/spreadsheets) | Spreadsheet | Free Tier |

## Deployment

Currently the deployment is manual. But the following are some improvements:

### AWS

- Terraform is not used. The only AWS resource is an S3 bucket.
- S3 bucket `yt-extractor` is created manually from the console. 

### Neon

- Make use of branches
- Install CLI
- Automate using API

### Railway

- Install CLI
- Automate Deployment using API

### Make.com

- Stop Automations via API
- Start Automations via API

## Issues

- Need a new Github repo.
- Dockefile must be optimized.
- Debug statement appear not strcutured in Railway log output.
- Need a tool to prepare for deployment:
    - Remove all bucket entries.
    - ~~Truncate database tables.~~
    - Delete Google Sheet rows.
    - Delete Notion Rows.
- ~~Observability? We have errors table in addition to Railway.~~
- ~~Make.com automations require variables to store channel ID and API KEY. It turned out there is something called scenario input that is useful in this case.~~
- ~~Add notes on the automations.~~
- ~~Add Reset Factory API?~~
- ~~Audio Split to generate transcription.~~
- ~~Video Summary in Arabic and English is no longer needed.~~
- ~~Video ID `E8yRq75_yBo` is being converted by the Cloud Convert to `E8yRq75yBo`!!!! This is solved by using CloudConvert file name instead of the video ID.~~
- Make.com automations must use both Google Sheets and Notion.

## Automations

These automations require a Youtube channel ID to operarte on and an API Key:

| Automation      | Description                       | Interval | Size |
|-----------------|-----------------------------------|----------|------|
| Pull            | Request yt videos be pulled from Youtube using API  | Every 24 hrs | -1 |
| Extract         | Request unextracted yt videos be extracted into S3   | Every 15 Minutes | 10 |
| Externalization | Export extracted videos to external sheets (Google and Notion)   | Every 30 mins  | 100 |
| Processing | Processes extracted and externalized videos to generate audio and transcription files  | Every 1 Hour  | 10 |
| Updation | Updates extracted, externalized and processed videos to set the latest video metrics: comments, views and likes  | Every 30 mins  | 100 |

## Risks

- Google YT API Key expiration in TEST mode. 
- Extractor tool constant updates may require periodic deployments.