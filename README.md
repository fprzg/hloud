# hloud
Simple Home Cloud

This is project meant as a very simple clone of consumer cloud storage services, such as Google Drive or Mediafire.

## API Design

### File Uploads
The api only supports one file per http request. It could support multiple files per request (as multpart), but my reasoning for not allowing that is the following:
1. The files are small (< 10mb). Making one request per file uploaded may be a big overhead because the files would upload very fast, but since the files do upload fast (because they are small), it may not be too harmful (and this project is not meant to store 100k files per request).
2. The files are big (>= 10mb). Then the overhead per file is almost neglectible, since the files will take a few seconds, even minutes. If we end up closing a connection due to an error, the penalization would be too big. We could also had a time connection limit error.
3. Chunk upload system. If the file is too big (>= 10mb), the file gets chunked and uploaded as such. This prevents us from having time connection exceeded errors (if the body of the request is too big). It could also be easily extended to prevent that specific type of error on slow connections.

## Frontend Design
