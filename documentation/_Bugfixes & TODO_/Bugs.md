# Bugs

## High

 - Have CRON script to restart server in case of catastrophic server failure
 - Cache assembly and part lists & refresh cache in background when the collection is updated
 - Dangerous deletion prevention
  - Allow for user, part, and assembly IDs to be changed in case something important that is deleted needs to be added back later
  - Show a warning when trying to delete
    - A user who has made orders in the past
    - A part that is used in any assemblies
  - Don't actually delete anything until *n* days later
  - Regularly archive databases and uploaded media
 - Encrypt database connection string somewhere other than `config.yaml`!
 - Strip leading "http[s]://" and trailing "/" from entries in `config.yaml`.
 - Dynamically serve images for everything from `/upload/UUID-GOES-HERE/` instead of storing the URL in the object
 - Use the DataTables API to gather `input` data instead of clearing the search fields with jquery.
 - Sort `GetAll` database returns by alphabetical order.

## Medium

 - Tidy up CSS
 - Ensure db.ImageSet()
   - Limits maximum file size to 8 mB [16 mB is max supported by MongoDB]
   - Converts images to highly compressed webp and a 1:1 aspect ratio, with a maximum size of 1000x1000 pixels.
   - May use a separate CDN
 - Make part / user / assembly item order consistent.
 - Change AuditLogOrder.Order to AuditLogOrder.ID for consistency
 - /assembly/ID doesn't return 404 on invalid IDs
 - The images in /assembly/ID float behind the parts table.
 - Load Packaging Methods from config

## Low

 - Create style guide
   - Variables are of the form Noun(Verb/Adjective) -- "TargetGet" rather than "GetTarget". This makes sorting code alphabetically easier.
   - Consistent capitalization.
   - Clean up badly-written functions to be more concise.
   - Consistent error and log messages - whether to use ending punctuation
 - Make "cost per unit" in the part creation page easier to use -- allow "$10.32" instead of "1032"
 - Sort part names on /assembly/ in alphabetical order.
