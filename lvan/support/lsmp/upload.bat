go build ../../cmd/mp
GnuZip\zip.exe -r lsmp.zip GnuZip lsmp.bat mp.exe
.\nexus-tool.exe upload -f lsmp.zip -g unity-plugin-lsmp -v 1.0.0 -r gforge-editor-tools --no_latest