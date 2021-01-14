# Future Dynamic Assembly Description

So, at some point we want to be able to describe any kind of field item without making a new creation / ordering page for each type (standard wall rough assembly, etc.)
They'll be described with YAML as follows:

# Assemblies

ID: The assembly's UUID.
Name: The name of the object
Category: What kind of standard assembly is this? Temporary Power? Wallrough?
Parts:
    2228cddf-6a17-4390-a7d5-c476a4e39f77: 1
    0093bb79-2b69-4247-bcec-1edb5159b616: 1
    bd6d8673-22f6-4aa7-891d-9dec51a55698: 2
BuildTime: 20
Media:
    Primary: The basic image for the assembly.
    CAD: A 3D model of the assembly.
    FooBar: Any other images that may be relevant.

# Parts

THIS IS ALREADY DONE -- SEE `/lib/db.Part` FOR A STRUCT DESCRIPTION.

# Categories

Categories require assemblies to have certain fields.

Name: The name of the category.
Description: What kinds of assemblies should be in this category?
Fields:
    - Bracket Style:
      - Floor Stand
      - Single Box Adjustable
      - Multi Box Adjustable
      - Single Box Nearest Stud
    - Raceway Style:
      - Conduit
      - MC Cable
      - Both
    - System:
      - Generic
      - Lighting
      - Fire
      - Power
    - Device:
      - Yes
      - No