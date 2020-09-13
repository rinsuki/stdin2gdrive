# stdin2gdrive

Upload stdin contents to Google Drive.


## Requirements

- Shared Drive (it needs G Suite)
- Service Account with Google Drive API enabled
- JSON Key File of Service Account

## How to use

### Upload from FFmpeg

```
ffmpeg -re -f lavfi -i testsrc2=s=640x360 -r 60 -t 60 -vcodec h264 -f flv - | stdin2gdrive --name testsrc2.1minute.flv --parent Folder_ID_Here --service-account /path/to/key.json
```
