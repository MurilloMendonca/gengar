local function endsWith(base, extension)
    assert(type(base) == "string", "Base parameter must be a string")
    assert(type(extension) == "string", "Extension parameter must be a string")
    local start_pos = #base - #extension + 1
    if start_pos < 1 then
        return false
    end
    return string.sub(base, start_pos) == extension
end

local function contains(table, element)
    assert(type(table) == "table")
    for _, value in ipairs(table) do
        if value == element then
            return true
        end
    end
    return false
end

local function trim(str)
    assert(type(str) == "string", "str parameter must be a string")
    return string.match(str, "^%s*(.-)%s*$")
end

local function stringContains(base, substring)
    assert(type(base) == "string", "Base parameter must be a string")
    assert(type(substring) == "string", "Substring parameter must be a string")
    return string.find(base, substring, 1, true) ~= nil
end

local function stringAppend(base, toAppend)
    assert(type(base) == "string", "Base parameter must be a string")
    assert(type(toAppend) == "string" or type(toAppend) == "table", "toAppend parameter must be a string or a table")
    if type(toAppend) == "string" then
        return base .. toAppend
    elseif type(toAppend) == "table" then
        for _, v in ipairs(toAppend) do
            base = base .. v
        end
    end
    return base
end

local function getNextWord(str, start)
    assert(type(str) == "string", "str parameter must be a string")
    assert(type(start) == "number", "start parameter must be a number")
    local finish = start
    for i = start + 1, #str do
        finish = finish + 1
        local c = string.sub(str, i, i)
        if c == " " or c == nil then
            return trim(string.sub(str, start, finish - 1)), finish
        end
    end
    return trim(string.sub(str, start, finish)), finish
end

local function isSrcFile(file)
    return endsWith(file, ".c") or endsWith(file, ".cpp")
end

local compile_commands = {}

local function generateCompileCommandsJson(path)
    path = path or "build/compile_commands.json"
    local file = io.open(path, "w")
    if file == nil then
        print("Failed to open compile_commands.json file for writing")
        return
    end
    file:write("[\n")
    for i, cmd in ipairs(compile_commands) do
        file:write("\t{\n")
        file:write("\t\t\"command\": \"" .. cmd.command .. "\",\n")
        file:write("\t\t\"file\": \"" .. cmd.file .. "\",\n")
        file:write("\t\t\"directory\": \"" .. cmd.directory .. "\"\n")
        if i == #compile_commands then
            file:write("\t}\n")
        else
            file:write("\t},\n")
        end
    end
    file:write("]\n")
    file:close()
end

local function runCmd(cmd, obj)
    if obj then
        compile_commands[#compile_commands + 1] = {
            command = cmd, file = obj.output, directory = os.getenv("PWD") }
    end

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

local function file_exists(name)
    assert(type(name) == "string", "name paramter must be a string")
    local f = io.open(name, "r")
    if f ~= nil then
        io.close(f)
        return true
    else
        return false
    end
end

local function isDirectory(path)
    assert(type(path) == "string", "path paramter must be a string")
    -- detph is equal to the number of "/" in the path string
    local depth = select(2, string.gsub(path, "/", ""))
    local cmd = "find " .. path .. " -maxdepth " .. depth .. " -type d"
    local handle = io.popen(cmd)
    if handle == nil then
        print("Failed to open handle")
        return false
    end
    local result = handle:read("*a")
    result = string.match(result, "^%s*(.-)%s*$")
    handle:close()
    if result == path then
        -- print(path, " is a directory")
        return true
    else
        -- print(path, " is not a directory")
        return false
    end
end


local function printIfNotNil(value, prefix)
    if value then
        if type(value) == "table" then
            print(prefix)
            for _, v in ipairs(value) do
                print("\t\t" .. v)
            end
        elseif type(value) == "boolean" then
            print(prefix .. tostring(value))
        else
            print(prefix .. value)
        end
    end
end

local function info(project)
    print("----- Project Info -----")
    printIfNotNil(project.name, "Name: ")
    printIfNotNil(project.version, "Version: ")
    printIfNotNil(project.description, "Description: ")
    printIfNotNil(project.compiler, "Compiler: ")
    printIfNotNil(project.flags, "Flags: ")
    printIfNotNil(project.generateCompileCommands, "Generate compile commands: ")
    printIfNotNil(project.output, "Output: ")

    if project.modules ~= nil then
        print("Modules: ")
        for _, module in ipairs(project.modules) do
            print("\t----- Module Info -----")
            printIfNotNil(module.name, "\tName: ")
            printIfNotNil(module.version, "\tVersion: ")
            printIfNotNil(module.sources, "\tSources: ")
            printIfNotNil(module.include, "\tInclude: ")
            printIfNotNil(module.static, "\tStatic: ")
            printIfNotNil(module.output, "\tOutput: ")
            printIfNotNil(module.executable, "\tExecutable: ")
            printIfNotNil(module.libraries, "\tLibraries: ")
        end
    else
        print("Modules: None")
    end
end

local function getSrcFilesFromPath(path)
    if not isDirectory(path) then
        return { path }
    end

    if not endsWith(path, "/") then
        path = path .. "/"
    end
    local dirFiles = io.popen("ls " .. path)
    if dirFiles == nil then
        print("Failed to open directory")
        return {}
    end
    local files = {}
    for file in dirFiles:lines() do
        if isSrcFile(file) then
            files[#files + 1] = path .. file
        end
    end
    return files
end

local function buildDynamicLibrary(baseCmd, module)
    print("----- Building dynamic library: " .. module.name .. " -----")
    local cmdForDotSo = baseCmd .. " -shared"
    for _, src in ipairs(module.sources) do
        local srcFiles = getSrcFilesFromPath(src)
        for _, srcFilePath in ipairs(srcFiles) do
            cmdForDotSo = cmdForDotSo .. " " .. srcFilePath .. ".o"
        end
    end
    cmdForDotSo = cmdForDotSo .. " -o " .. module.output .. ".so"
    if module.version then
        cmdForDotSo = cmdForDotSo .. "." .. module.version
        runCmd(cmdForDotSo)
        cmdForDotSo = "ln -sf " ..
            os.getenv("PWD") .. "/" .. module.output .. ".so." .. module.version .. " " .. module.output .. ".so"
        runCmd(cmdForDotSo)
        local major = string.match(module.version, "^(%d+)")
        cmdForDotSo = "ln -sf " ..
            os.getenv("PWD") ..
            "/" .. module.output .. ".so." .. module.version .. " " .. module.output .. ".so." .. major
    end
    runCmd(cmdForDotSo)
end

local function buildStaticLibrary(baseCmd, module)
    print("----- Building static library: " .. module.name .. " -----")
    _ = baseCmd
    local cmdForDotA = "ar rcs " .. module.output .. ".a"
    for _, src in ipairs(module.sources) do
        local srcFiles = getSrcFilesFromPath(src)
        for _, srcFilePath in ipairs(srcFiles) do
            cmdForDotA = cmdForDotA .. " " .. srcFilePath .. ".o"
        end
    end
    runCmd(cmdForDotA)
end

local function buildDotOForFile(baseCmd, srcFilePath, module)
    local cmdForDotO = baseCmd .. " -c " .. srcFilePath
    cmdForDotO = cmdForDotO .. " -o " .. srcFilePath .. ".o"
    if module.static then
        cmdForDotO = cmdForDotO .. " -static"
    end
    runCmd(cmdForDotO, module)
end

local function buildObjectFiles(baseCmd, module)
    print("----- Building object files for: " .. module.name .. " -----")
    if module.preBuildCommands then
        return module.preBuildCommands(runCmd, module)
    end
    if module.sources then
        for _, src in ipairs(module.sources) do
            local srcFiles = getSrcFilesFromPath(src)
            for _, srcFilePath in ipairs(srcFiles) do
                buildDotOForFile(baseCmd, srcFilePath, module)
            end
        end
    end
end

local function buildExecutable(baseCmd, module)
    print("----- Building executable: " .. module.name .. " -----")
    local cmdForDotExe = baseCmd
    for _, src in ipairs(module.sources) do
        local srcFiles = getSrcFilesFromPath(src)
        for _, srcFilePath in ipairs(srcFiles) do
            cmdForDotExe = cmdForDotExe .. " " .. srcFilePath .. ".o"
        end
    end
    cmdForDotExe = cmdForDotExe .. " -o " .. module.output
    if module.libraries then
        cmdForDotExe = cmdForDotExe .. " -Lbuild"
        for _, lib in ipairs(module.libraries) do
            cmdForDotExe = cmdForDotExe .. " -l" .. lib
        end
    end
    if module.static then
        cmdForDotExe = cmdForDotExe .. " -static"
    else
        cmdForDotExe = cmdForDotExe .. " -Wl,-rpath=" .. os.getenv("PWD") .. "/build"
    end

    runCmd(cmdForDotExe, module)
end

local function buildModule(cmd, module)
    if module.buildCommands then
        print("----- Running " .. module.name .. " custom build commands -----")
        module.buildCommands(runCmd, module)
    elseif module.executable then
        buildExecutable(cmd, module)
    elseif module.static then
        buildStaticLibrary(cmd, module)
    else
        buildDynamicLibrary(cmd, module)
    end
end

local function getLinksAndIncludesFromHaunter()
    if not file_exists("haunter.lua") then
        print("Haunter file not found")
        return "", ""
    end
    local haunter = require("haunter")
    local links = {}
    local includes = {}
    for _, lib in ipairs(haunter) do
        if lib.includePath then
            includes[#includes + 1] = lib.includePath
        end
        if lib.libPath then
            links[#links + 1] = lib.libPath
        end
    end
    return links, includes
end

local function addIncludesToCmd(cmd, includes, prefix)
    prefix = prefix or ""
    for _, include in ipairs(includes) do
        cmd = cmd .. " -I" .. prefix .. include
    end
    return cmd
end

local function addLibPathsToCmd(cmd, includes, prefix)
    prefix = prefix or ""
    for _, include in ipairs(includes) do
        cmd = cmd .. " -L" .. prefix .. include
    end
    return cmd
end

local function isLibrary(lib)
    return stringContains(lib, ".a") or stringContains(lib, ".so")
end

local function getAllLibrariesFromPath(path)
    local cmd = "ls " .. path
    local handle = io.popen(cmd)
    if handle == nil then
        print("Failed to open handle")
        return {}
    end
    local result = handle:read("*a")
    handle:close()
    local libs = {}
    for lib in string.gmatch(result, "%S+") do
        if isLibrary(lib) then
            libs[#libs + 1] = lib
        end
    end
    return libs
end

local function symlinkDepsLibsToBuild(links)
    for _, link in ipairs(links) do
        local libs = getAllLibrariesFromPath("dependencies/" .. link)
        for _, lib in ipairs(libs) do
            local cmd = "ln -sf " .. os.getenv("PWD") .. "/dependencies/" .. link .. lib .. " build/" .. lib
            runCmd(cmd)
        end
    end
end


local function getBaseModuleCmd(module, project, haunterIncludes)
    local moduleCmd = ""
    if module.compiler then
        moduleCmd = module.compiler .. " "
    else
        moduleCmd = project.compiler .. " "
    end
    if module.flags then
        moduleCmd = moduleCmd .. module.flags
    else
        moduleCmd = moduleCmd .. project.flags
    end
    -- Add includes from haunter again
    if haunterIncludes and haunterIncludes ~= "" then
        moduleCmd = addIncludesToCmd(moduleCmd, haunterIncludes, "dependencies/")
    end
    return moduleCmd
end

local function build(project)
    mkdirIfNotExists("build")
    local links, includes = getLinksAndIncludesFromHaunter()
    if links and links ~= "" then
        symlinkDepsLibsToBuild(links)
    end

    if project.modules ~= nil then
        for _, module in ipairs(project.modules) do
            local moduleCmd = getBaseModuleCmd(module, project, includes)
            local includeSet = {}
            if module.include then
                for _, include in ipairs(module.include) do
                    if not contains(includeSet, include) then
                        includeSet[#includeSet + 1] = include
                        moduleCmd = moduleCmd .. " -I" .. include
                    end
                end
            end

            buildObjectFiles(moduleCmd, module)
        end
        for _, module in ipairs(project.modules) do
            if not module.executable then
                buildModule(getBaseModuleCmd(module, project, includes), module)
            end
        end
        for _, module in ipairs(project.modules) do
            if module.executable then
                buildModule(getBaseModuleCmd(module, project, includes), module)
            end
        end
    end
    if project.generateCompileCommands then
        generateCompileCommandsJson()
    end
end

local function clean(project)
    local cmd = "rm -rf build/"
    runCmd(cmd)
    if project then
        if project.modules then
            for _, module in ipairs(project.modules) do
                if module.sources then
                    for _, src in ipairs(module.sources) do
                        local srcFiles = getSrcFilesFromPath(src)
                        for _, srcFilePath in ipairs(srcFiles) do
                            cmd = "rm -f " .. srcFilePath .. ".o"
                            runCmd(cmd)
                        end
                    end
                end
            end
        end
    end
end



local shOutputFile = nil
local function genSh(project)
    shOutputFile = io.open("generated.sh", "w")
    if shOutputFile == nil then
        print("Failed to open generated.sh file for writing")
        return
    end
    runCmd = function(cmd, obj)
        if obj then
            compile_commands[#compile_commands + 1] = {
                command = cmd, file = obj.output, directory = os.getenv("PWD") }
        end
        shOutputFile:write(cmd .. "\n")
    end

    build(project)
    shOutputFile:close()
end
local function mapModulesByType(modules)
    local dynLibs = {}
    local staticLibs = {}
    local executables = {}
    for _, mod in ipairs(modules) do
        if mod.executable then
            executables[#executables + 1] = mod
        elseif mod.static then
            staticLibs[#staticLibs + 1] = mod
        else
            dynLibs[#dynLibs + 1] = mod
        end
    end
    return dynLibs, staticLibs, executables
end

local makefile = nil
local function genMakefile(project)
    makefile = io.open("Makefile", "w")
    if makefile == nil then
        print("Failed to open Makefile file for writing")
        return
    end
    makefile:write("CC = " .. project.compiler .. "\n")
    makefile:write("CFLAGS = " .. project.flags .. "\n")
    -- local baseCmd = project.compiler .. " " .. project.flags .. " "

    makefile:write("\nall: buildDir dynamicLibs staticLibs executables")
    makefile:write("\nbuildDir:\n\tmkdir -p build")

    local dynLibs, staticLibs, execs = mapModulesByType(project.modules)
    makefile:write("\ndynamicLibs: ")
    if dynLibs then
        for _, mod in ipairs(dynLibs) do
            if mod.output then
                makefile:write(mod.output .. ".so ")
            end
        end
    end

    makefile:write("\nstaticLibs: ")
    if staticLibs then
        for _, mod in ipairs(staticLibs) do
            if mod.output then
                makefile:write(mod.output .. ".a ")
            end
        end
    end

    makefile:write("\nexecutables: ")
    if execs then
        for _, mod in ipairs(execs) do
            if mod.output then
                makefile:write(mod.output .. " ")
            end
        end
    end


    runCmd = function(cmd, obj)
        if obj then
            compile_commands[#compile_commands + 1] = {
                command = cmd, file = obj.output, directory = os.getenv("PWD") }
        end
        if stringContains(cmd, "-o ") then
            local _, start = string.find(cmd, "-o ")
            local out = getNextWord(cmd, start)
            makefile:write("\n" .. out .. ":")
            if obj and stringContains(out, obj.output) then
                for _, src in ipairs(obj.sources) do
                    local srcFiles = getSrcFilesFromPath(src)
                    for _, srcFilePath in ipairs(srcFiles) do
                        makefile:write(" " .. srcFilePath .. ".o")
                    end
                end
            end
            makefile:write("\n\t" .. cmd)
        elseif stringContains(cmd, "ar rcs ") then
            local _, start = string.find(cmd, "ar rcs ")
            local out = getNextWord(cmd, start)
            makefile:write("\n" .. out .. ":\n\t" .. cmd)
        end
    end
    build(project)
    makefile:close()
end


local function show()
    print([[⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⡔⠋⠣⡄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⡏⠀⠀⢠⠇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠑⢢⣶⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣤⠔⠒⢦⣄⣀⣀⡴⠋⠁⠉⢧⠀⠀⠀⠀⠀⠀⠀⢀⡴⠦⣄⠀⠀⠀⠀⠀⣄⡴⠊⠓⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢰⡇⠀⠀⠀⠀⠀⠀⠉⠀⠀⠀⣀⡾⠀⠀⠀⣀⣤⠤⠖⠋⠀⠀⠀⠓⢦⡀⠀⢸⡆⠀⠀⠀⠈⢹⡄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⡤⠖⠚⠀⠀⠀⠀⠀⠀⠀⠀⠀⠠⣟⠉⣀⡀⠀⣸⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⠢⠎⠀⠀⠀⠀⠀⣾⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⢇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠉⠉⠉⠉⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣹⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣀⣸⠇⠀⠀⠀⠀⢀⣀⣠⡤⠤⠶⠶⠶⠶⠶⠤⢤⣀⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⡶⠶⣄⠀⠀⣠⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⣰⠒⢲⡄⠀⠀⠀⢠⡞⠁⠀⠀⠀⠀⣠⠴⠚⠉⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠙⠲⢤⣀⠀⠀⠀⠀⠀⠠⡟⠁⠀⠈⠙⠙⣁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⣟⠀⢸⡇⠀⣦⠀⣴⠃⠀⠀⢀⡴⠋⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣴⣿⣿⣿⣷⣦⡀⠀⠀⠉⠳⣄⡀⠀⠀⠀⠘⣆⠀⣀⣤⠞⠉⠑⢦⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠈⠉⣿⠋⠙⠛⡞⠁⠀⢀⡼⠋⠀⣠⣤⣤⣀⡀⠀⠀⠀⠀⠀⣀⣸⣿⣿⣿⣿⣿⣿⣶⡄⠀⠀⠀⠈⠻⣄⠀⠀⠀⠀⠉⠁⠀⠀⠀⠀⠈⠳⢄⠀⠀⠀⢀⡖⠚⠷⡀
⠀⢠⡇⠀⠀⣰⠻⡄⣼⠟⠀⠀⢀⣿⣿⣿⣿⣿⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣤⡀⠀⠀⠈⢳⡄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢨⡇⠀⠀⢘⡆⢠⡴⠃
⢠⠏⠀⠀⢰⠃⠀⢻⡋⠀⣠⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣾⣧⠀⢳⡄⠀⠀⠀⠀⠀⠀⠀⠀⠀⣼⠀⠀⢰⡎⠀⠘⠧⣄
⠈⠑⣄⠀⡏⠀⠀⠀⢷⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠛⠉⣿⣿⣿⣷⠀⢳⠀⠀⠀⠀⠀⠀⠀⣠⡴⠋⠀⠀⠸⣇⡀⠀⠀⡎
⠀⠀⢵⢀⡇⠀⠀⠀⠈⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠛⠁⠀⠀⠀⢸⣿⣿⣿⣧⠈⣇⠀⠀⠀⠀⠀⣰⠋⠀⠀⠀⠀⠀⠀⠙⠒⠚⠁
⠀⠀⠈⢹⠁⠀⠀⠀⠀⠈⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠛⠁⠀⠀⠀⠀⠀⠀⠈⣿⣿⣿⣿⠀⢻⠀⠀⠀⠀⠀⣿⠀⠀⣴⠛⠉⠛⠉⠳⣄⠀⠀
⠀⠀⠀⠸⡄⠀⠀⠀⠀⠀⠈⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠟⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢻⣿⣿⣿⡇⢸⠀⠀⠀⠀⠀⠸⢆⣀⡿⠀⠀⠀⠀⠀⢈⡇⠀
⠀⠀⠀⠀⡇⠀⠀⠀⠀⠀⢀⡎⢻⡿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⡇⢸⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠸⡆⠀
⠀⠀⠀⢀⣿⡀⠀⠀⠀⠀⠈⠃⠀⣿⣾⣿⣿⣿⣯⣿⠟⠁⠀⣴⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⣿⣿⣿⣿⠁⣾⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠻⡄
⠀⠀⡶⠋⠘⢧⠀⠀⠀⠀⠀⠀⢠⣿⣿⣿⣿⣿⣿⣿⠀⠀⠀⠙⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣾⣿⣿⣿⡇⢠⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⡰⠃
⠀⠀⢦⠀⠀⠈⢷⣀⠀⠀⠀⢠⣾⣿⣿⣿⣿⣿⣿⣿⣧⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⣾⣿⣿⣿⠟⢀⡿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣀⣀⡼⠁⠀
⠀⠀⢸⡄⠀⠀⠀⠙⣷⣶⢾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣿⣦⡀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣴⣿⡿⠛⠛⠁⢠⡾⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⡾⠋⠈⠉⠀⠀⠀
⠀⠀⠀⠈⠉⠻⡄⠀⣿⢉⡷⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣦⣤⣄⣠⣤⣴⣶⣿⣿⠟⠄⠀⠀⣰⠏⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣸⠁⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⢸⠀⠻⠞⠀⠘⢿⣍⠛⠻⠿⠿⣿⣿⣿⣿⣿⠿⢿⡛⠋⣿⣿⣿⣿⠿⠋⠁⠀⠀⣠⠞⠁⠀⠀⠀⠀⠀⠀⠀⣠⡴⠛⠒⠒⠚⠁⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠸⡄⠀⠀⢀⣀⡈⠻⣷⣦⡀⠀⠀⠀⠀⠀⠀⠀⣀⣷⣴⣿⡿⠋⠀⠀⠀⢀⡤⠞⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⡿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠈⠓⠒⠋⠈⠉⠢⡌⠙⠿⢿⣗⣲⠶⠶⠶⠿⠿⠿⠟⠛⢁⣀⣤⠶⠚⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢰⣇⠀⣀⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⠀⠀⠀⠈⠉⠉⠛⠛⠒⠒⠛⠛⠉⠉⠁⠀⠀⠀⠀⣀⣀⠀⠀⠀⠀⠀⠀⠀⠀⢀⡀⢼⡟⠁⠈⢻⠄⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣇⠀⠀⠀⠀⠀⠀⠀⠀⢰⡶⠶⢦⠀⠀⠀⠀⠀⠀⣸⠇⠀⠙⡇⠀⠀⠀⠀⣴⠞⠁⡀⠀⠹⠤⠴⠊⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠙⠲⢄⠀⠀⠀⠀⠀⠈⢳⠀⠰⣤⣀⣠⣀⡠⠞⠁⠀⠀⠀⠙⠲⢤⣀⡴⠋⠀⠸⣍⡷⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠑⣤⠤⠖⠒⠲⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀]])
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
    print("No command provided, trying to run 'info' by default")
    local project = require("gastly")
    info(project)
    return 0
end

if (command == "info") then
    local luaFile = getArg()
    local project = returnRequireIfExist(luaFile, "gastly")
    info(project)
elseif (command == "show") then
    show()
elseif (command == "build") then
    local luaFile = getArg()
    local project = returnRequireIfExist(luaFile, "gastly")
    build(project)
elseif (command == "gen-make") then
    local luaFile = getArg()
    local project = returnRequireIfExist(luaFile, "gastly")
    genMakefile(project)
elseif (command == "gen-sh") then
    local luaFile = getArg()
    local project = returnRequireIfExist(luaFile, "gastly")
    genSh(project)
elseif (command == "clean") then
    local luaFile = getArg()
    local project = returnRequireIfExist(luaFile, "gastly")
    clean(project)
else
    print("Command not recognized")
    return 1
end
