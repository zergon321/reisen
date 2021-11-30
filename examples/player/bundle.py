import os, fileinput, shutil

dllDir = "C:/msys64/mingw64/bin"

bundleDirPath = os.path.abspath("bundle")
os.makedirs(bundleDirPath, exist_ok=True)

for dllName in fileinput.input():
    dllName = dllName.strip()
    dllPath = os.path.join(dllDir, dllName)
    dllBundlePath = os.path.join(bundleDirPath, dllName)

    shutil.copyfile(dllPath, dllBundlePath)

shutil.copyfile("player.exe", "bundle/player.exe")