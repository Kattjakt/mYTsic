# mYTsic
A basic wrapper for ffmpeg and youtube-dl created with music downloading in mind. Contains features such as user-uploads & playlist downloading, automatic title/artist tagging (using ID3v2) and Cover Images. 


### Notes
mYTsic is extremely sensitive and will discard any video/file which doesn't have a fully parsable title. As of now, the metadata is parsed out of the video title ("artist - title"), using the hypen as a separator. This means that any video containing less or more than two hyphens will be discarded. This also applies to videos where the cover image has a resolution less than 600x600. 
### Todo's
- ?