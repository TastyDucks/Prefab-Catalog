# 1.0 Release Features

 - [x] Login
 - [ ] MONGODB MIGRATION!!!
   - [x] Users
   - [ ] Parts
   - [ ] Assemblies
   - [ ] Orders
 - [ ] Order mailing
 - [ ] Audit log
   - [ ] Administrative record
   - [ ] Order statistics

# Misc Bugfixes

## High

- Have CRON script to restart server in case of catastrophic server failure
- Dangerous deletion prevention
  - Allow for user, part, and assembly IDs to be changed in case something important that is deleted needs to be added back later
  - Show a warning when trying to delete:
    - A user who has made orders in the past
    - A part that is used in any assemblies
  - Don't actually delete anything until *n* days later
  - Regularly archive databases and uploaded media
- Encrypt database connection string somewhere other than `config.yaml`!
- Dynamically serve images for everything from `/upload/UUID-GOES-HERE/` instead of storing the URL in the object

## Medium

 - Tidy up CSS
 - Ensure db.ImageSet()
   - Limits maximum file size to 8 mB [16 mB is max supported by MongoDB]
   - Converts images to highly compressed webp and a 1:1 aspect ratio, with a maximum size of 1000x1000 pixels.
   - May use a seperate CDN
- Make part / user / assembly item order consistent.
- Change AuditLogOrder.Order to AuditLogOrder.ID for consistency
- /assembly/ID doesn't return 404 on invalid IDs

## Low

 - Create style guide
   - Variables are of the form Noun(Verb/Adjective) -- "TargetGet" rather than "GetTarget". This makes sorting code alphabetically easier.
   - Consistent capitalization.
   - Clean up badly-written functions to be more concise.
   - Consistent error and log messages - whether to use ending punctuation
 - Force browser to clear image cache when updating images for users, parts, or assemblies.
 - Make "cost per unit" in the part creation page easier to use -- allow "$10.32" instead of "1032"

# Future Features

 - Integrate user accounts with existing UAC system
 - Load collections and object descriptions from `config.yaml`, auto-generate routes and table displays for them! Maximum expandability just by editing a config file! (Users will be hardcoded though)
 - Keep a listing of on-site parts and list parts to-be-purchased
   - Train a neural network to predict future part needs and automatically order supplies as needed (with review by order dept)
 - Email SysAdmin if there is an `Error` or `Fatal` log entry

# Build history

 - 0.1: SQLite base (pre-release)
 - 0.2: Change to MongoDB (pre-release)
 - 0.3: Order Completion (pre-release)
 - 1.0: Alpha Launch