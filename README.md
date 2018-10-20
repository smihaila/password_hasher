# password_hasher
CloudJump Password Hasher Microservice

1. Project structure:

- docs/ dir: Documentation.
- src/  dir: All the source code.
- build.cmd: Microsoft Windows-specific build script. It will create the bin/ dir when run from the project's root dir.

2. Building & running:
a. Windows:
   i) Execute in a first cmd.exe shell:
     - cd <project_root_dir>
     - build.cmd
     - bin\password_hasher.exe
   ii) Execute in a second cmd.exe shell (or use PostMan):
     - curl -X POST -d 'password=some_password' http://localhost:8080/hash
     - curl http://localhost:8080/stats
     - curl http://localhost:8080/shutdown
b. Linux: tbd
c. Mac OSX: tbd
