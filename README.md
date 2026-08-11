# gcal-popup
This project takes freeform text as input and turns it into a Google Calendar event. 

## Requirements
- We only need to support a single user
- We shouldn't write an incorrect event to the calendar. If text is ambiguous, ask clarifying questions
- We will cap clarifying questions at a few rounds, then fall back to best guess if still unclear

## Design
- Hotkey opens a popup where I can input text
- Text will go to an LLM which will extract title / date / time / duration
- If text is ambiguous, LLM will ask a clarifying question in that same popup 
    - Questions are capped at 3 rounds. After that, the LLM will go with its best guess and user can confirm/fix manually
- Once confirmed, event is written directly to gcal via API
