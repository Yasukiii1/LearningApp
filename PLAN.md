# LearningApp Plan

## Current project location

C:\Users\brain\OneDrive\Documents\learningproject

## Current project tree

learningproject/
├── builds/
│   └── LearningApp.exe
├── main.go
├── go.mod
├── build.bat
├── .gitignore
└── PLAN.md

## Local-only file

C:\Users\brain\OneDrive\Documents\learningproject\New Text Document.txt

This file is the personal directory/checkpoint note and is intentionally ignored by Git.

## Current state

- GitHub repository: Yasukiii1/LearningApp
- Branch: main
- Checkpoint 0: initial executable
- Checkpoint 1: source code, build script, and basic white/yellow UI
- Sidebar base implementation is now in source code.
- Local Go build environment needs to be available in CMD.

## UI reference

Preferred text style reference:

𝐋𝐞𝐚𝐫𝐧𝐢𝐧𝐠𝐀𝐩𝐩

This is Unicode Mathematical Bold styling, not the name of a normal UI font. Final UI typography has not been chosen yet.

### External visual reference

- Dribbble shot: "Sidebar navigation for Dashboard" by Tran Mau Tri Tam.
- LearningApp may use the shot as close visual inspiration for sidebar composition, hierarchy, spacing, compact controls, and general interaction patterns.
- LearningApp will use its own text, branding, icons, colors, dimensions, and detailed visual treatment rather than making a pixel-for-pixel copy.
- The bottom-right image-upload area from the reference is not part of LearningApp's planned sidebar.

## Accepted UI planning

### Sidebar

- The application will have a collapsible sidebar.
- The sidebar will use a clean, light grey/white visual direction rather than the earlier yellow palette.
- The top-left area will contain the LearningApp app mark.
- The top navigation/search area will contain a compact note-name search field and the sidebar toggle.
- The search field will be used to search saved note names.
- The sidebar will initially contain at least five blank smooth square icon slots/buttons as placeholders.
- The purposes, names, and final icons for those five or more slots are not decided yet.
- Slightly above the bottom of the sidebar there will be a profile row containing a profile icon/avatar and the user's name, following the compact structure of the visual reference.
- At the bottom of the sidebar there will be a simple Settings item/button with a dedicated small icon area.
- Settings is a placeholder for now and will do nothing initially.
- The reference's compact notification/utility-control idea can be used as inspiration for a future LearningApp utility control.
- Hover/meaning indicators and their animations are planned for a later stage after the base UI is complete.
- Sidebar interactions and exact icon meanings will be decided later.

### Application window

- The app uses a normal Windows application window.
- The window is resizable and supports the standard minimize, maximize/restore, and close controls supplied by the operating system.
- The application launches maximized while retaining the standard Windows title-bar controls.

### Visual direction

- Keep the existing overall sidebar composition rather than replacing it with a different navigation model.
- Use whitespace, restrained borders, rounded corners, and light neutral surfaces.
- Blank square sidebar icons remain intentionally blank until their purposes are decided.
- The overall feel should be compact, minimal, modern, and suitable for a learning/knowledge application.

## Accepted initial subjects

These are the only subjects planned for the initial version:

1. English First Language
2. Chemistry
3. Physics
4. Add Maths
5. Maths
6. Malay
7. Chinese

## Planning status

Sidebar planning is currently the first major planning task. Chapter structure, learning workflows, visual tools, storage, and AI generation systems will be planned later.
