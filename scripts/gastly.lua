return {
    name = "MyProject",
    version = "0.1.0",
    description = "A simple project template",
    compiler = "cc",
    flags = "-std=c99",
    generateCompileCommands = true,
    modules = {
        {
            name = "Dynamic lib foo",
            version = "11.1.0",
            sources = {
                "src/foo/",
            },
            include = { "include/foo/"},
            static = false,
            output = "build/libfoo",
            preBuildCommands = function(execCmdFunc, thisModule)
                return true
            end,
            buildCommands= function (execCmdFunc,thisModule)
                execCmdFunc("gcc -std=c99 -Iinclude/foo src/foo/someCode.c src/foo/someOtherCode.c -shared -o build/libfoo.so",thisModule)
                return true
            end,
        },
        {
            name = "Final Dyn Executable",
            version = "0.1.0",
            sources = {
                "src",
            },
            include = {
                "include/",
                "include/foo/",
            },
            libraries = {
                "foo",
            },
            static = false,
            output = "build/dyn-exec",
            executable = true,
        },
        {
            name = "Static lib foo",
            version = "0.1.0",
            sources = {
                "src/foo/someCode.c",
                "src/foo/someOtherCode.c",
            },
            include = {
                "include/foo/",
            },
            static = true,
            output = "build/libfoo",
        },
        {
            name = "Final Executable",
            version = "0.1.0",
            sources = {
                "src",
            },
            include = {
                "include/",
                "include/foo/",
            },
            libraries = {
                "foo",
            },
            static = true,
            output = "build/exec",
            executable = true,
        },
    },
    output = {"build/myproject"},
}
