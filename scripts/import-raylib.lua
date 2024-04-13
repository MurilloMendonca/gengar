local raylib = {
    url= "https://github.com/raysan5/raylib/releases/download/5.0/raylib-5.0_linux_amd64.tar.gz",
    name = "raylib",
    includePath="raylib/include",
    libPath="raylib/lib"
}

local tinyLib = {
    url= "https://github.com/MurilloMendonca/tiny-libs/archive/refs/heads/master.zip",
    name = "tiny-libs",
    includePath="tiny-libs/",
    libPath="tiny-libs/",
    buildCommands= function(localDir)
        local code = os.execute("mv "..localDir.."/tiny-libs-master/* "..localDir)
        if (code) then
            code = os.execute("rm -rf "..localDir.."/tiny-libs-master/")
            if code then
                return true
            else
                return false
            end
        else
            return false
        end

    end,
}

local tinyLibFromGit = {
    url= "https://github.com/MurilloMendonca/tiny-libs",
    name = "tiny-libs-git",
    isGit = true,
    includePath="tiny-libs-git/",
    libPath="tiny-libs-git/",
}

local libs = {tinyLibFromGit, raylib, tinyLib}
return libs
