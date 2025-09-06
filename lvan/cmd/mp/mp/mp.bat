echo 拷贝 mp.exe 到插件目录
set plugin_dir=C:\Users\lvan\AppData\Roaming\g.editor.electron.v1\otherPlugin\UnityTransformLocalPriData\1.0.0\
xcopy /Y mp.exe %plugin_dir%
explorer %plugin_dir%