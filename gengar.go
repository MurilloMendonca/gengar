package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type BuildSystem string
type CppStandard string
type Compiler string
type TestingFramework string

const (
	Makefile   BuildSystem      = "Makefile"
	CMake      BuildSystem      = "CMake"
	BuildSh    BuildSystem      = "build.sh"
	Premake5   BuildSystem      = "Premake5"
	GCC        Compiler         = "GCC"
	Clang      Compiler         = "Clang"
	MSVC       Compiler         = "MSVC"
	Cpp17      CppStandard      = "c++17"
	Cpp20      CppStandard      = "c++20"
	Cpp23      CppStandard      = "c++23"
	GoogleTest TestingFramework = "Google Test"
	Catch2     TestingFramework = "Catch2"
	BoostTest  TestingFramework = "Boost.Test"
)

type model struct {
	currentStep      int
	cursor           int
	projectName      string
	buildSystem      BuildSystem
	cppStandard      CppStandard
	compiler         Compiler
	testingFramework TestingFramework
	projectStructure []string // Customizable directory structure
	optionsList      map[int][]string
	steps            []string // To navigate between different setup steps
}

func initialModel() model {

	// Create a list of options for each step
	optionsList := map[int][]string{
		0: {"Premake5", "Makefile", "CMake", "build.sh", "None"},
		1: {"c++17", "c++20", "c++23"},
		2: {"GCC", "Clang", "MSVC"},
		3: {"Google Test", "None"},
	}
	return model{
		currentStep: 0,
		steps: []string{"Build System",
			"C++ Standard",
			"Compiler",
			"Testing Framework",
		},
		optionsList: optionsList,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter":
			m.handleSelection()
			if m.currentStep < len(m.steps)-1 {
				m.currentStep++
			} else {
				return m, tea.Quit
			}
			return m, nil
		case "up", "k":
			// Move the cursor up
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			// Move the cursor down
			if m.cursor < len(m.optionsList[m.currentStep])-1 {
				m.cursor++
			}
			return m, nil

		case "n":
			// Navigate to the next step
			if m.currentStep < len(m.steps)-1 {
				m.currentStep++
			}
			return m, nil

		case "p":
			// Navigate to the previous step
			if m.currentStep > 0 {
				m.currentStep--
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *model) handleSelection() {

	switch m.steps[m.currentStep] {
	case "Build System":
		m.buildSystem = BuildSystem(m.optionsList[m.currentStep][m.cursor]) // Update the model with the selected build system
	case "C++ Standard":
		m.cppStandard = CppStandard(m.optionsList[m.currentStep][m.cursor]) // Update the model with the selected C++ standard
	case "Compiler":
		m.compiler = Compiler(m.optionsList[m.currentStep][m.cursor]) // Update the model with the selected compiler
	case "Testing Framework":
		m.testingFramework = TestingFramework(m.optionsList[m.currentStep][m.cursor]) // Update the model with the selected testing framework
	}
}

func (m model) View() string {
	// Display the current step and instructions
	stepTitle := "Step: " + m.steps[m.currentStep]
	instructions := "Use arrow keys to select, enter to confirm, 'n' for next, 'p' for previous, 'q' to quit."

	// Render the list
	listStr := ""
	for i, option := range m.optionsList[m.currentStep] {
		if i == m.cursor {
			listStr += ">" // Add a cursor to the selected option
		} else {
			listStr += " "
		}
		listStr += " " + option + "\n"
	}

	return fmt.Sprintf("%s\n\n%s\n\n%s", stepTitle, listStr, instructions)
}

func createDirectoryStructure(basePath string) error {
	dirs := []string{
		"build",
		filepath.Join("include", "foo"),
		filepath.Join("src", "foo"),
	}

	for _, dir := range dirs {
		path := filepath.Join(basePath, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}
	return nil
}

func generateExampleFiles(basePath string) error {
	// Base files to create
	files := map[string]string{
		filepath.Join("src", "main.cpp"): `#include <iostream>
#include "foo/foo.hpp"

int main() {
    std::cout << "Hello, World!" << std::endl;
    std::cout << "40 + 2 = " << foo::bar(40, 2) << std::endl;
    return 0;
}
`,
		filepath.Join("src", "foo", "foo.cpp"): `#include "foo/foo.hpp"

namespace foo {
    int bar(int a, int b) {
        return a + b;
    }
}
`,
		filepath.Join("include", "foo", "foo.hpp"): `#pragma once

namespace foo {
    int bar(int a, int b);
}
`,
	}
	for filePath, content := range files {
		fullPath := filepath.Join(basePath, filePath)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %v", fullPath, err)
		}
	}
	return nil

}

func generateCMakeLists(m model) error {
	fmt.Println("Ignoring Compiler for CMake build system")
	basePath := filepath.Join(".", m.projectName)
	content := `cmake_minimum_required(VERSION 3.10)
project(` + m.projectName + ` VERSION 1.0)

# Specify the C++ standard
set(CMAKE_CXX_STANDARD ` + strings.Split(string(m.cppStandard), "c++")[1] + `)
set(CMAKE_CXX_STANDARD_REQUIRED True)

# Include directories
include_directories(include)

# Create a shared library 'libfoo.so' from 'src/foo/foo.cpp'
add_library(foo SHARED src/foo/foo.cpp)

# Create the main program executable
add_executable(main src/main.cpp)

# Link the main program with the foo library
target_link_libraries(main foo)
`

	switch m.testingFramework {
	case "Google Test":
		generateTestFiles(m)
		content += `# Add the google test framework as a build dependency
include(FetchContent)
FetchContent_Declare(
  googletest
  URL https://github.com/google/googletest/archive/03597a01ee50ed33e9dfd640b249b4be3799d395.zip
)
set(gtest_force_shared_crt ON CACHE BOOL "" FORCE)
FetchContent_MakeAvailable(googletest)

enable_testing()

# Create a test executable for the foo library
add_executable(
  foo-tests
  tests/unit-tests/foo-tests.cpp
)
target_link_libraries(
  foo-tests
  GTest::gtest_main
  foo
)

include(GoogleTest)
gtest_discover_tests(foo-tests)

`
	default:
	}

	if err := os.WriteFile(filepath.Join(basePath, "CMakeLists.txt"), []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create file CMakeLists.txt: %v", err)
	}
	return nil
}

func generateMakefile(m model) error {
	basePath := filepath.Join(".", m.projectName)
	compilerStr := ""
	switch m.compiler {
	case "GCC":
		compilerStr = "g++"
	case "Clang":
		compilerStr = "clang++"
	case "MSVC":
		compilerStr = "cl"
	default:
		compilerStr = "g++"
	}
	content := `includePath := include
srcPath := src
binPath := build/bin
objPath := build/obj
libPath := build/lib

CC := ` + compilerStr + `
CFLAGS := -I$(includePath) -fPIC -std=` + string(m.cppStandard) + `
LDFLAGS := -L$(libPath) -Wl,-rpath,$(libPath)

# Target names
EXECUTABLE := $(binPath)/main
LIBRARY := $(libPath)/libfoo.so
MAIN_OBJ := $(objPath)/main.o
FOO_OBJ := $(objPath)/foo.o

all: directories $(LIBRARY) $(EXECUTABLE)

directories:
	mkdir -p $(binPath) $(objPath) $(libPath)

$(EXECUTABLE): $(MAIN_OBJ)
	$(CC) $^ -o $@ $(LDFLAGS) -lfoo

$(MAIN_OBJ): $(srcPath)/main.cpp
	$(CC) $(CFLAGS) -c $< -o $@

$(LIBRARY): $(FOO_OBJ)
	$(CC) -shared $^ -o $@

$(FOO_OBJ): $(srcPath)/foo/foo.cpp
	$(CC) $(CFLAGS) -c $< -o $@

clean:
	rm -rf $(binPath) $(objPath) $(libPath)

`

	switch m.testingFramework {
	case "Google Test":
		generateTestFiles(m)
		content += `# tests
testsPath := tests/unit-tests
testsBinPath := build/tests

TEST_EXECUTABLE := $(testsBinPath)/unit-tests

test-dir:
	mkdir -p $(testsBinPath)

test: directories test-dir $(TEST_EXECUTABLE)
	$(TEST_EXECUTABLE)

$(TEST_EXECUTABLE): $(testsPath)/foo-tests.cpp $(FOO_OBJ)
	$(CC) $(CFLAGS) $^ -o $@ $(LDFLAGS) -lfoo -lgtest -lgtest_main -pthread

clean-test:
	rm -rf $(testsBinPath)

cleanall: clean-test clean
`
	default:
	}

	if err := os.WriteFile(filepath.Join(basePath, "Makefile"), []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create file Makefile: %v", err)
	}

	return nil
}

func generateBuildSh(m model) error {
	basePath := filepath.Join(".", m.projectName)
	compilerStr := ""
	switch m.compiler {
	case "GCC":
		compilerStr = "g++"
	case "Clang":
		compilerStr = "clang++"
	case "MSVC":
		compilerStr = "cl"
	default:
		compilerStr = "g++"
	}
	content := `#!/bin/bash
# A simple build script
` + compilerStr + ` -std=` + string(m.cppStandard) + ` -Iinclude -o build/main src/main.cpp src/foo/foo.cpp
#!/bin/bash
# A simple build script

if [[ "$1" == "clean" ]]; then
  rm -rf build/*
  exit 0
fi

if [ ! -d "build" ]; then
  mkdir build
fi

if [[ "$1" == "shared" ]]; then
` + compilerStr + ` -std=` + string(m.cppStandard) + ` -Iinclude -c src/foo/foo.cpp -o build/foo.o
` + compilerStr + ` -std=` + string(m.cppStandard) + ` -shared -o build/libfoo.so build/foo.o
` + compilerStr + ` -std=` + string(m.cppStandard) + ` -Iinclude -o build/main src/main.cpp -Lbuild -lfoo -Wl,-rpath,./build
  exit 0
fi

if [[ "$1" == "static" ]]; then
` + compilerStr + ` -std=` + string(m.cppStandard) + ` -Iinclude -c src/foo/foo.cpp -o build/foo.o
  ar rcs build/libfoo.a build/foo.o
` + compilerStr + ` -std=` + string(m.cppStandard) + ` -Iinclude -o build/main src/main.cpp -Lbuild -lfoo
  exit 0
fi

# Default build step if none of the above conditions are met
` + compilerStr + ` -std=` + string(m.cppStandard) + ` -Iinclude -o build/main src/main.cpp src/foo/foo.cpp

`

	switch m.testingFramework {
	case "Google Test":
		generateTestFiles(m)
		content += `# tests
if [[ "$1" == "tests" ]]; then
` + compilerStr + ` -std=` + string(m.cppStandard) + ` -Iinclude -Itests/unit-tests -o build/foo-tests tests/unit-tests/foo-tests.cpp src/foo/foo.cpp -lgtest -lgtest_main -pthread
  ./build/foo-tests
  exit 0
fi

`
	default:
	}
	if err := os.WriteFile(filepath.Join(basePath, "build.sh"), []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create file build.sh: %v", err)
	}

	return nil

}

func generatePremake5Lua(m model) error {
	fmt.Println("Ignoring Compiler for Premake5 build system")
	basePath := filepath.Join(".", m.projectName)
	content := `workspace "` + m.projectName + `"
   configurations { "Debug", "Release" }
   architecture "x86_64"
   startproject "MainApp"

   flags
   {
       "MultiProcessorCompile"
   }

   -- Global settings for all configurations
   filter "system:linux"
      cppdialect "` + strings.ToUpper(string(m.cppStandard)) + `"
      buildoptions { "-fPIC" }

project "Foo"
   kind "SharedLib"
   language "C++"
   targetdir "bin/%{cfg.buildcfg}"
   targetname "foo"

   files { "src/foo/**.cpp", "include/foo/**.h" }
   includedirs { "include" }

   filter "configurations:Debug"
      defines { "DEBUG" }
      symbols "On"

   filter "configurations:Release"
      defines { "NDEBUG" }
      optimize "On"

project "MainApp"
   kind "ConsoleApp"
   language "C++"
   targetdir "bin/%{cfg.buildcfg}"

   files { "src/main.cpp" }
   includedirs { "include" }

   links { "Foo" }

   filter "configurations:Debug"
      defines { "DEBUG" }
      symbols "On"

   filter "configurations:Release"
      defines { "NDEBUG" }
      optimize "On"
`

	if err := os.WriteFile(filepath.Join(basePath, "premake5.lua"), []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create file premake5.lua: %v", err)
	}

	return nil
}
func createBuildSolution(m model) error {
	err := error(nil)
	switch m.buildSystem {
	case "CMake":
		err = generateCMakeLists(m)
	case "Makefile":
		err = generateMakefile(m)
	case "build.sh":
		err = generateBuildSh(m)
	case "Premake5":
		err = generatePremake5Lua(m)
	default:
		err = generatePremake5Lua(m)
	}
	if err != nil {
		return fmt.Errorf("failed to create build solution: %v", err)
	}

	return nil
}

func createProjectStructure(m model) error {
	basePath := filepath.Join(".", m.projectName)
	err := createDirectoryStructure(basePath)
	if err != nil {
		return fmt.Errorf("failed to create directory structure: %v", err)
	}
	err = generateExampleFiles(basePath)
	if err != nil {
		return fmt.Errorf("failed to generate example files: %v", err)
	}
	err = createBuildSolution(m)
	if err != nil {
		return fmt.Errorf("failed to create build solution: %v", err)
	}

	return nil
}

func generateTestFiles(m model) error {
	basePath := filepath.Join(".", m.projectName)
	content := `#include "gtest/gtest.h"
#include "foo/foo.hpp"

// Test case for the bar function
TEST(SumTest, HandlesPositiveInput) {
    EXPECT_EQ(42, foo::bar(40, 2));
}

TEST(SumTest, HandlesNegativeInput) {
    EXPECT_EQ(-1, foo::bar(-3, 2));
}

TEST(SumTest, HandlesZeroInput) {
    EXPECT_EQ(0, foo::bar(0, 0));
}
`

	testsDir := filepath.Join(basePath, "tests", "unit-tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", testsDir, err)
	}

	testFilePath := filepath.Join(testsDir, "foo-tests.cpp")
	if err := os.WriteFile(testFilePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create file %s: %v", testFilePath, err)
	}

	return nil
}
func initNewProject(projectName string) {
	m := initialModel()
	p := tea.NewProgram(m)
	output, err := p.Run()
	if err != nil {
		fmt.Println(err)
	}
	m = output.(model)
	m.projectName = projectName
	err = createProjectStructure(m)
	if err != nil {
		fmt.Println(err)
	}
	err = createBuildSolution(m)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Project created successfully")
}

func showUsage() {
	fmt.Println("Usage: gengar [init] <project-name>")
}

func main() {
	projectName := ""
	commandOption := ""
	if len(os.Args) > 1 {
		commandOption = os.Args[1]
	} else {
		showUsage()
		os.Exit(1)
	}
	switch commandOption {

	case "init":
		if len(os.Args) > 2 {
			projectName = os.Args[2]
		} else {
			fmt.Println("Please provide a project name")
			os.Exit(1)
		}

		initNewProject(projectName)

	case "help":
		showUsage()
	default:
		showUsage()
	}

}
