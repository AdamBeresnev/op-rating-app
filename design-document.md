# Op rating web app v1
## Abstract
This is a web app to help organize anime op rating events. 
The web app needs to simplify the process by:
- taking in a list of links
- generating editable entities from said list
- compiling entities into an editable tournament bracket
- let user view tournament bracket in bracket overview
- let user view tournament bracket in match view
- let user open corresponding links in match view
- let user pick a winner in match view
- fill subsequent tournament placing based on match winner and loser
## Components
### Executable program
The web app must be an executable that upon execution starts hosting a local website and opens said website.
The executable reads and/or creates tournament data (bracket entities, their attributes and current placings) in a seperate folder.
### Main menu
The main menu of the website displays all currently created tournaments and a button for new tournament creation.
- Selecting a created tournament opens the bracket overview with corresponding data.
- Pressing "New tournament" buttons opens tournament creation view.
### Tournament creation view
This view consists of a text box for tournament name, a text area where user can insert links from "animethemes.moe" and "youtube.com" separated by new lines and a tournament creation button.
- Expected link amount is 20, but the tournament creation process should support creation of smaller or bigger brackets.
- A number of entities created from inserted links should be visible to the user.
After the user inserts all links and presses the tournament creation button they are redirected to the bracket overview.
### Bracket overview
All tournament entities are displayed in their current positions and matchups. At the top there are buttons to "Change tournament settings" and "Start/Continue tournament".
- Entities are represented by their names and tournament bracket position numbers. Pressing on an entity opens a pop-up(?) to edit entity name, position, image url or undo its .
- "Start/Continue tournament" button is accessible only when all entities have distinct positions, conflicts are outlined in red. 
- "Change tournament settings" opens a pop-up(?) to edit tournament name and type (single/double elimination).
- "Start/Continue tournament" button opens match view.
### Match view
Match view displays a single tournament bracket position where two entities are competing. A button is present to return to bracket view.
- Entities are shown side by side (left and right).
- Entity image and information is displayed.
- Each entity has a button to open entity link in a new tab.
- Each entity has a button to select it as a winner
- Winning and losing entity are assigned new positions in the tournament according to tournament logic and type.
- Entity position history is kept so that it can be undone.
Upon selecting a winner another bracket position is loaded in match view.
### Podium view
Upon completing tournament top 4-8 positions are displayed in order of their results.
- Exiting podium view returns user to bracket view where a new "View podium" button is visible.
## Requirements
### High priority
- Editable tournament bracket creation from 20 entities
- Double elimination tournament creation view where barebones editable entities are created from 20 links.
- Image fetching from a user specified image url.
- Match view where two entities are displayed, their links can be accessed and a winner can be picked.
- Tournament information saving and parsing in a seperate file/file folder.
### Medium priority
- Tournament bracket can be created from a variable number of entities
- Entities can be created from a variable number of links 
- Entity names are extracted from animethemes.moe links
- Entity images are extracted from animethemes.moe links
- Tournament bracket type can be changed to single elimination
- Podium view
### Low priority
- Video embeding from links youtube.com links
- Video embeding from links animethemes.moe links
- Podium view can be saved as an image
- Entity names are extracted from youtube.com links
- Entity images are extracted from youtube.com links