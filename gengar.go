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
type DependencyManager string
type DocumentationTool string
type CICDTool string
type CodeFormatter string
type License string

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
		0: {"Premake5","Makefile", "CMake", "build.sh"},
		1: {"c++17", "c++20", "c++23"},
		2: {"GCC", "Clang", "MSVC"},
		//3: {"Google Test", "Catch2", "Boost.Test"},
	}
	return model{
		currentStep: 0,
		steps:       []string{"Build System", 
                            "C++ Standard", 
                            "Compiler", 
                        //    "Testing Framework"
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
	// case "Testing Framework":
	// 	m.testingFramework = TestingFramework(m.optionsList[m.currentStep][m.cursor]) // Update the model with the selected testing framework
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

func createProjectStructure(m model) error {
	basePath := filepath.Join(".", m.projectName)
	// Directories to create
	dirs := []string{
		"build",
		filepath.Join("include", "sum"),
		filepath.Join("src", "sum"),
	}

	for _, dir := range dirs {
		path := filepath.Join(basePath, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", path, err)
		}
	}

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

	// Base files to create
	files := map[string]string{
		filepath.Join("src", "main.cpp"): `#include <iostream>
#include "sum/sum.hpp"

int main() {
    std::cout << "Hello, World!" << std::endl;
    std::cout << "40 + 2 = " << sum::add(40, 2) << std::endl;
    return 0;
}
`,
		filepath.Join("src", "sum", "sum.cpp"): `#include "sum/sum.hpp"

namespace sum {
    int add(int a, int b) {
        return a + b;
    }
}
`,
		filepath.Join("include", "sum", "sum.hpp"): `#pragma once

namespace sum {
    int add(int a, int b);
}
`,
	}

	// Append Build System specific file
	switch m.buildSystem {
	case "CMake":
		files[filepath.Join("CMakeLists.txt")] = `cmake_minimum_required(VERSION 3.10)
project(` + m.projectName + ` VERSION 1.0)

# Specify the C++ standard
set(CMAKE_CXX_STANDARD ` + strings.Split(string(m.cppStandard), "c++")[0] + `)
set(CMAKE_CXX_STANDARD_REQUIRED True)

# Include directories
include_directories(include)

# Create a shared library 'libsum.so' from 'src/sum/sum.cpp'
add_library(sum SHARED src/sum/sum.cpp)

# Create the main program executable
add_executable(main src/main.cpp)

# Link the main program with the sum library
target_link_libraries(main sum)
`
	case "Makefile":
		files[filepath.Join("Makefile")] = `includePath := include
srcPath := src
binPath := build/bin
objPath := build/obj
libPath := build/lib

CC := ` + compilerStr + `
CFLAGS := -I$(includePath) -fPIC -std=` + string(m.cppStandard) + `
LDFLAGS := -L$(libPath) -Wl,-rpath,$(libPath)

# Target names
EXECUTABLE := $(binPath)/main
LIBRARY := $(libPath)/libsum.so
MAIN_OBJ := $(objPath)/main.o
SUM_OBJ := $(objPath)/sum.o

all: directories $(LIBRARY) $(EXECUTABLE)

directories:
	mkdir -p $(binPath) $(objPath) $(libPath)

$(EXECUTABLE): $(MAIN_OBJ)
	$(CC) $^ -o $@ $(LDFLAGS) -lsum

$(MAIN_OBJ): $(srcPath)/main.cpp
	$(CC) $(CFLAGS) -c $< -o $@

$(LIBRARY): $(SUM_OBJ)
	$(CC) -shared $^ -o $@

$(SUM_OBJ): $(srcPath)/sum/sum.cpp
	$(CC) $(CFLAGS) -c $< -o $@

clean:
	rm -rf $(binPath) $(objPath) $(libPath)

`
	case "build.sh":
		files[filepath.Join("build.sh")] = `#!/bin/bash
# A simple build script
` + compilerStr + ` -std=` + string(m.cppStandard) + ` -Iinclude -o main src/main.cpp src/sum/sum.cpp
`
    case "Premake5":
        files[filepath.Join("premake5.lua")] = `workspace "` + m.projectName + `"
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

project "Sum"
   kind "SharedLib"
   language "C++"
   targetdir "bin/%{cfg.buildcfg}"
   targetname "sum"

   files { "src/sum/**.cpp", "include/sum/**.h" }
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

   links { "Sum" }

   filter "configurations:Debug"
      defines { "DEBUG" }
      symbols "On"

   filter "configurations:Release"
      defines { "NDEBUG" }
      optimize "On"

`
	default:
		// Handle default case or unsupported build systems
	}

	for filePath, content := range files {
		fullPath := filepath.Join(basePath, filePath)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %v", fullPath, err)
		}
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

	case "help": showUsage()
	default:
		showUsage()
	}

}
