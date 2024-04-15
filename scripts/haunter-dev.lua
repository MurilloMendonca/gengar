local function runCmd(cmd, obj)
    print(cmd)
    local code, _, _ = os.execute(cmd)
    if (code) then
        -- print("Command executed successfully")
        return 0
    else
        print("Failed to execute command")
        return 1
    end
end
local function mkdirIfNotExists(path)
    assert(type(path) == "string", "path parameter must be a string")
    local code, _, _ = runCmd("mkdir -p " .. path)
    if (code) then
        print("Created dir: " .. path)
    else
        print("Failed to create directory")
        return 1
    end
end
local function stringContains(base, substring)
    assert(type(base) == "string", "Base parameter must be a string")
    assert(type(substring) == "string", "Substring parameter must be a string")
    return string.find(base, substring, 1, true) ~= nil
end
local function endsWith(base, extension)
    assert(type(base) == "string", "Base parameter must be a string")
    assert(type(extension) == "string", "Extension parameter must be a string")
    local start_pos = #base - #extension + 1
    if start_pos < 1 then
        return false
    end
    return string.sub(base, start_pos) == extension
end
local function getFileNameFromUrl(url)
    local size = 1
    while size < #url do
        local start = #url - size
        if stringContains(string.sub(url, start, #url), "/") then
            return string.sub(url, start + 1, #url)
        end
        size = size + 1
    end
    return url
end

local function stripExtention(filename)
    local size = 1
    while size < #filename do
        local start = #filename - size
        if stringContains(string.sub(filename, start, #filename), ".") then
            return string.sub(filename, 1, start - 1)
        end
        size = size + 1
    end
    return filename
end

local function downloadFile(url, path)
    assert(type(url) == "string", "url must be a string")
    path = path or ("dependencies/" .. getFileNameFromUrl(url))
    runCmd("wget " .. url .. " -O " .. path)
    return path
end


local function gitClone(repoUrl, path)
    assert(type(repoUrl) == "string", "repoUrl must be a string")
    path = path or ("dependencies/" .. getFileNameFromUrl(repoUrl))
    runCmd("git clone --depth=1 " .. repoUrl .. " " .. path)
    return path
end

local function unzipFile(path, dest)
    assert(type(path) == "string", "path must be a string")
    dest = dest or "dependencies"
    runCmd("unzip " .. path .. " -d " .. dest)
end

local function untarFile(path, dest)
    assert(type(path) == "string", "path must be a string")
    dest = dest or "dependencies"
    runCmd("tar -xvzf " .. path .. " -C " .. dest .. " --strip-components=1")
end
local function folderAlreadyExists(path)
    assert(type(path) == "string", "path must be a string")
    local code, _, _ = runCmd("ls " .. path)
    return code == 0
end
local function import(lib)
    assert(type(lib) == "table", "lib parameter must be a table")
    assert(type(lib.url) == "string", "url parameter must be a string")
    assert(type(lib.name) == "string", "name parameter must be a string")
    assert(type(lib.includePath) == "string", "includePath parameter must be a string")
    assert(type(lib.libPath) == "string", "libPath parameter must be a string")
    if folderAlreadyExists("dependencies/" .. lib.includePath) and
        folderAlreadyExists("dependencies/"..lib.libPath)then
        print("Library already exists")
        return 0
    end
    local localDir = "dependencies/" .. lib.name
    mkdirIfNotExists(localDir)

    -- Fist Step: Download the file
    if lib.downloadCommands then
        if lib.downloadCommands(localDir) then
            print("Download commands executed successfully")
        else
            print("Failed to execute download commands")
            return 1
        end
    elseif lib.isGit then
        gitClone(lib.url, localDir)
    elseif endsWith(lib.url, ".zip") then
        downloadFile(lib.url, localDir .. ".zip")
    elseif endsWith(lib.url, ".tar.gz") then
        downloadFile(lib.url, localDir .. ".tar.gz")
    else
        print("Unsupported file format, just downloading")
        downloadFile(lib.url, localDir .."/" ..getFileNameFromUrl(lib.url))
    end

    -- Second Step: Unpack/move the files to the correct location
    if lib.unpackCommands then
        if lib.unpackCommands(localDir) then
            print("Unpack commands executed successfully")
        else
            print("Failed to execute unpack commands")
            return 1
        end
    elseif endsWith(lib.url, ".zip") then
        unzipFile(localDir .. ".zip", localDir)
        runCmd("rm " .. localDir .. ".zip")
    elseif endsWith(lib.url, ".tar.gz") then
        untarFile(localDir .. ".tar.gz", localDir)
        runCmd("rm " .. localDir .. ".tar.gz")
    end


    -- Third Step: Execute build commands
    if lib.buildCommands then
        if lib.buildCommands(localDir) then
            print("Build commands executed successfully")
        else
            print("Failed to execute build commands")
            return 1
        end
    end
    return 0
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

local function clean(project)
    runCmd("rm -rf dependencies")
end
local function returnRequireIfExist(name, fallback)
    if name == nil or name == "" then
        print("No lua file provided, searching for fallback")
        return require(fallback)
    else
        if endsWith(name, ".lua") then
            name = string.sub(name, 1, #name - 4)
        end
        return require(name)
    end
end
local index = 0
if getArg == nil then
    getArg = function()
        index = index + 1
        return arg[index]
    end
end

local command = getArg()
if (command == nil) then
    local libs = require("haunter")
    for _, lib in ipairs(libs) do
        import(lib)
    end
end

if (command == "info") then
    local luaFile = getArg()
    local project = returnRequireIfExist(luaFile, "haunter")
    info(project)
elseif (command == "get") then
    local luaFile = getArg()
    local projects = returnRequireIfExist(luaFile, "haunter")
    for _, project in ipairs(projects) do
        import(project)
    end
elseif (command == "clean") then
    clean()
else
    print("Command not recognized")
    return 1
end
