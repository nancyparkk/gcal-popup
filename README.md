# gcal-popup

You can type freeform like `lunch with x on monday around 1ish` and if it's too
vague to work with, this will prompt a question. 

## How it works

The menu bar app runs a small server on the machine. The text goes to Gemini, which
pulls out the title, date, time and how long it should be. Once you confirm, it writes the event
to Google Calendar.

The server doesn't remember anything between requests. The
browser keeps track of the conversation and sends the whole thing each time. 

## Setup

In the Cloud Console, make a project and turn on the Calendar API. Create
an OAuth client ID, pick Desktop app, and save the file it gives you as credentials.json in the
project folder. Add your own Google account as a test user.

Then get Gemini key from AI Studio. Put it in a file called .env:

    GEMINI_API_KEY=your-key-here

log in:

    go run ./cmd/auth

Look at the address bar, copy the long code out of the URL, and
paste it back into the terminal to save your login automatically. 

## Launch

    go run .

A calendar icon shows up in the menu bar and follow as prompted!


## Known problems

Google logged me out every week since this app hasn't been through its review process. 
When adding events starts failing, run go run ./cmd/auth again. Could be avoided by
adding a privacy policy but would require publishing. 