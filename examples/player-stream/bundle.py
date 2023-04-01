import os, fileinput, shutil

bundleDirPath = os.path.abspath("bundle")
os.makedirs(bundleDirPath, exist_ok=True)

for dllInfo in fileinput.input():
    dllInfo = dllInfo.strip()
    dllInfoParts = dllInfo.split(sep=" ")
    dllName = dllInfoParts[0]
    dllPath = dllInfoParts[2]
    dllBundlePath = os.path.join(bundleDirPath, dllName)

    if dllPath.startswith("/mingw64/bin"):
        dllPath = os.path.join("C:/msys64", dllPath[1:])
        shutil.copyfile(dllPath, dllBundlePath)

shutil.copyfile("player.exe", "bundle/player.exe")