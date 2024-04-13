local function endsWith(base, extension)
    assert(type(base) == "string", "Base parameter must be a string")
    assert(type(extension) == "string", "Extension parameter must be a string")
    local start_pos = #base - #extension + 1
    if start_pos < 1 then
        return false
    end
    return string.sub(base, start_pos) == extension
end
local function import(lib)
    local code
    local dependencies_dir = "dependencies/" .. lib.name

    code, _, _ = os.execute("mkdir -p " .. dependencies_dir)
    if (code) then
        print("Created dir: " .. dependencies_dir)
    else
        print("Failed to create directory")
        return 1
    end
    if (lib.isGit) then
        print("Cloning repository...")
        code, _, _ = os.execute("git clone " .. lib.url .. " " .. dependencies_dir)
        if (code) then
            print("Cloned repository!")
        else
            print("Failed to clone repository")
            return 1
        end
        print("Cleaning...")
        code, _, _ = os.execute("rm -rf " .. dependencies_dir .. "/.git")
        if (code) then
            print("Cleaned repository!")
        else
            print("Failed to clean repository")
            return 1
        end
        if lib.buildCommands ~= nil then
            print("Building...")
            if lib.buildCommands(dependencies_dir) then
                print("Build Successful")
            else
                print("Failed to Build")
                return 1
            end
        end
        print("Library ready with:\n\tInclude path: " .. lib.includePath .. "\n\tLib path: " .. lib.libPath)
        return
    else
        print("Downloading file...")

        local package_url = lib.url
        code, _, _ = os.execute("wget " .. package_url .. " -O " .. dependencies_dir .. "/temp.tar.gz")

        if (code) then
            print("Downloaded file!")
        else
            print("Failed to download file")
            return 1
        end

        print("Unpacking...")

        if (endsWith(lib.url, "tar.gz")) then
            code, _, _ = os.execute("tar -xvzf " ..
                dependencies_dir .. "/temp.tar.gz -C " .. dependencies_dir .. " --strip-components=1")
        elseif (endsWith(lib.url, "zip")) then
            code, _, _ = os.execute("unzip " ..
                dependencies_dir .. "/temp.tar.gz -d " .. dependencies_dir)

            if (code) then
                print("Unpacked file!")
            else
                print("Failed to unpack file")
                return 1
            end
        else
            print("File format not supported!!")
            return 1
        end


        if (code) then
            print("Unpacked file!")
        else
            print("Failed to unpack file")
            return 1
        end

        print("Cleaning...")


        code, _, _ = os.execute("rm " .. dependencies_dir .. "/temp.tar.gz")

        if lib.buildCommands ~= nil then
            print("Building...")
            if lib.buildCommands(dependencies_dir) then
                print("Build Successful")
            else
                print("Failed to Build")
                return 1
            end
        end
    end

    print("Library ready with:\n\tInclude path: " .. lib.includePath .. "\n\tLib path: " .. lib.libPath)
end

local function info(libs)
    print("Haunter is a simple tool to manage dependencies in C projects")
    print("Usage: haunter-dev.lua get [luaFile]")
    print("luaFile: A lua file with the dependencies to be imported")
    if libs then
        print("Available libraries:")
        for _, lib in ipairs(libs) do
            print("\t" .. lib.name)
        end
    end
end

local command = getArg()
if (command == nil) then
    local libs = require("haunter")
    for _, lib in ipairs(libs) do
        import(lib)
    end
end

if (command == "get") then
    local luaFile = getArg()
    if (luaFile == nil) then
        print("No lua file provided, searching for gastly.lua")
        local libs = require("haunter")
        for _, lib in ipairs(libs) do
            import(lib)
        end
    else
        local libs = require(luaFile)
        for _, lib in ipairs(libs) do
            import(lib)
        end
    end
elseif (command == "info") then
    local luaFile = getArg()
    if (luaFile == nil) then
        print("No lua file provided, searching for gastly.lua")
        local libs = require("haunter")
        info(libs)
    else
        local libs = require(luaFile)
        info(libs)
    end
else
    print("Command not recognized")
    return 1
end
