package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	tea "github.com/charmbracelet/bubbletea"
    "github.com/yuin/gopher-lua"
)

type BuildSystem string
type CppStandard string
type Compiler string
type TestingFramework string

type model struct {
	currentStep      int
	cursor           int
	ProjectName      string
	buildSystem      string
	CppStandard      string
	compiler         string
	testingFramework string
	projectStructure []string // Customizable directory structure
	optionsList      map[int][]string
	steps            []string // To navigate between different setup steps
}

var templatePath = filepath.Join(".", "templates")
var scriptsPath = filepath.Join(".", "scripts")

func initialModel() model {

	// Create a list of options for each step
	optionsList := map[int][]string{
		0: {"Premake5", "Makefile", "CMake", "build.sh", "None"},
		1: {"c++17", "c++20", "c++23"},
		2: {"GCC", "Clang", "MSVC"},
		3: {"Google Test", "Catch2", "None"},
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
                m.cursor = 0
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
		m.buildSystem = (m.optionsList[m.currentStep][m.cursor]) // Update the model with the selected build system
	case "C++ Standard":
		m.CppStandard = (m.optionsList[m.currentStep][m.cursor]) // Update the model with the selected C++ standard
	case "Compiler":
		m.compiler = (m.optionsList[m.currentStep][m.cursor]) // Update the model with the selected compiler
	case "Testing Framework":
		m.testingFramework = (m.optionsList[m.currentStep][m.cursor]) // Update the model with the selected testing framework
	}
}

func (m model) View() string {
     // Display the current step and instructions with color
    stepTitle := fmt.Sprintf("\033[38;5;147mStep: %s\033[0m", m.steps[m.currentStep])

    instructions := "\033[38;5;210mUse arrow keys to select, enter to confirm, 'n' for next, 'p' for previous, 'q' to quit.\033[0m"

    // Render the list with color
    listStr := ""
    for i, option := range m.optionsList[m.currentStep] {
        if i == m.cursor {
            listStr += "\033[38;5;99m> " // Add a cursor to the selected option with color
        } else {
            listStr += "\033[38;5;99m  "
        }
        listStr += option + "\033[0m\n" // Reset color at the end of each option
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
	m.ProjectName = strings.ReplaceAll(m.ProjectName, " ", "_")
	m.CppStandard = strings.Split(string(m.CppStandard), "c++")[1]
	basePath := filepath.Join(".", m.ProjectName)
	tmplPath := filepath.Join(templatePath, "CMakeLists.tmpl")
	tmpl, err := template.New("CMakeLists.tmpl").ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse CMakeLists.tmpl: %v", err)
	}
	var content = new(strings.Builder)
	err = tmpl.Execute(content, m)
	if err != nil {
		return fmt.Errorf("failed to execute CMakeLists.tmpl: %v", err)
	}

	switch m.testingFramework {
	case "Google Test":
		generateGTestFiles(m)
		testContent := new(strings.Builder)
		tmplPath = filepath.Join(templatePath, "CMakeListsGTest.tmpl")
		tmpl, err = template.New("CMakeListsGTest.tmpl").ParseFiles(tmplPath)
		if err != nil {
			return fmt.Errorf("failed to parse CMakeListsGTest.tmpl: %v", err)
		}
		err = tmpl.Execute(testContent, m)
		if err != nil {
			return fmt.Errorf("failed to execute CMakeListsGTest.tmpl: %v", err)
		}
		content.WriteString(testContent.String())
    case "Catch2":
        generateCatch2Files(m)
        testContent := new(strings.Builder)
        tmplPath = filepath.Join(templatePath, "CMakeListsCatch2.tmpl")
        tmpl, err = template.New("CMakeListsCatch2.tmpl").ParseFiles(tmplPath)
        if err != nil {
            return fmt.Errorf("failed to parse CMakeListsCatch2.tmpl: %v", err)
        }
        err = tmpl.Execute(testContent, m)
        if err != nil {
            return fmt.Errorf("failed to execute CMakeListsCatch2.tmpl: %v", err)
        }
        content.WriteString(testContent.String())
	default:
	}

	if err := os.WriteFile(filepath.Join(basePath, "CMakeLists.txt"), []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to create file CMakeLists.txt: %v", err)
	}

	return nil
}

func generateMakefile(m model) error {
	basePath := filepath.Join(".", m.ProjectName)
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
	tmplPath := filepath.Join(templatePath, "Makefile.tmpl")
	tmpl, err := template.New("Makefile.tmpl").ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse Makefile.tmpl: %v", err)
	}
	var content = new(strings.Builder)
	type Data struct {
		ProjectName string
		CompileStr  string
		CppStandard string
	}
	data := Data{ProjectName: m.ProjectName, CompileStr: compilerStr, CppStandard: string(m.CppStandard)}
	err = tmpl.Execute(content, data)
	if err != nil {
		return fmt.Errorf("failed to execute Makefile.tmpl: %v", err)
	}

	switch m.testingFramework {
	case "Google Test":
		generateGTestFiles(m)
		testContent := new(strings.Builder)
		tmplPath = filepath.Join(templatePath, "MakefileGTest.tmpl")
		tmpl, err = template.New("MakefileGTest.tmpl").ParseFiles(tmplPath)
		if err != nil {
			return fmt.Errorf("failed to parse MakefileGTest.tmpl: %v", err)
		}
		err = tmpl.Execute(testContent, data)
		if err != nil {
			return fmt.Errorf("failed to execute MakefileGTest.tmpl: %v", err)
		}
		content.WriteString(testContent.String())
    case "Catch2":
        generateCatch2Files(m)
        testContent := new(strings.Builder)
        tmplPath = filepath.Join(templatePath, "MakefileCatch2.tmpl")
        tmpl, err = template.New("MakefileCatch2.tmpl").ParseFiles(tmplPath)
        if err != nil {
            return fmt.Errorf("failed to parse MakefileCatch2.tmpl: %v", err)
        }
        err = tmpl.Execute(testContent, data)
        if err != nil {
            return fmt.Errorf("failed to execute MakefileCatch2.tmpl: %v", err)
        }
        content.WriteString(testContent.String())
	default:
	}

	if err := os.WriteFile(filepath.Join(basePath, "Makefile"), []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to create file Makefile: %v", err)
	}

	return nil
}

func generateBuildSh(m model) error {
	basePath := filepath.Join(".", m.ProjectName)
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
	type Data struct {
		ProjectName string
		CompileStr  string
		CppStandard string
	}
	data := Data{ProjectName: m.ProjectName, CompileStr: compilerStr, CppStandard: string(m.CppStandard)}
	tmplPath := filepath.Join(templatePath, "build.sh.tmpl")
	tmpl, err := template.New("build.sh.tmpl").ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse build.sh.tmpl: %v", err)
	}
	var content = new(strings.Builder)
	err = tmpl.Execute(content, data)
	if err != nil {
		return fmt.Errorf("failed to execute build.sh.tmpl: %v", err)
	}

	switch m.testingFramework {
	case "Google Test":
		generateGTestFiles(m)
		testContent := new(strings.Builder)
		tmplPath = filepath.Join(templatePath, "buildGTest.tmpl")
		tmpl, err = template.New("buildGTest.tmpl").ParseFiles(tmplPath)
		if err != nil {
			return fmt.Errorf("failed to parse buildGTest.tmpl: %v", err)
		}
		err = tmpl.Execute(testContent, data)
		if err != nil {
			return fmt.Errorf("failed to execute buildGTest.tmpl: %v", err)
		}
		content.WriteString(testContent.String())
    case "Catch2":
        generateCatch2Files(m)
        testContent := new(strings.Builder)
        tmplPath = filepath.Join(templatePath, "buildCatch2.tmpl")
        tmpl, err = template.New("buildCatch2.tmpl").ParseFiles(tmplPath)
        if err != nil {
            return fmt.Errorf("failed to parse buildCatch2.tmpl: %v", err)
        }
        err = tmpl.Execute(testContent, data)
        if err != nil {
            return fmt.Errorf("failed to execute buildCatch2.tmpl: %v", err)
        }
        content.WriteString(testContent.String())
	default:
	}

	if err := os.WriteFile(filepath.Join(basePath, "build.sh"), []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to create file build.sh: %v", err)
	}

	return nil

}

func generatePremake5Lua(m model) error {
	fmt.Println("Ignoring Compiler for Premake5 build system")
	basePath := filepath.Join(".", m.ProjectName)
	tmplPath := filepath.Join(templatePath, "premake5.lua.tmpl")
	tmpl, err := template.New("premake5.lua.tmpl").ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse premake5.lua.tmpl: %v", err)
	}
	var content = new(strings.Builder)
	type Data struct {
		ProjectName string
		CppStandard string
		Compiler    string
	}

	data := Data{ProjectName: m.ProjectName,
		CppStandard: strings.ToUpper(string(m.CppStandard)),
		Compiler:    m.compiler}
	err = tmpl.Execute(content, data)

	switch m.testingFramework {
	case "Google Test":
		generateGTestFiles(m)
		testContent := new(strings.Builder)
		tmplPath = filepath.Join(templatePath, "premake5GTest.tmpl")
		tmpl, err = template.New("premake5GTest.tmpl").ParseFiles(tmplPath)
		if err != nil {
			return fmt.Errorf("failed to parse premake5GTest.tmpl: %v", err)
		}
		err = tmpl.Execute(testContent, data)
		if err != nil {
			return fmt.Errorf("failed to execute premake5GTest.tmpl: %v", err)
		}
		content.WriteString(testContent.String())
    case "Catch2":
        generateCatch2Files(m)
        testContent := new(strings.Builder)
        tmplPath = filepath.Join(templatePath, "premake5Catch2.tmpl")
        tmpl, err = template.New("premake5Catch2.tmpl").ParseFiles(tmplPath)
        if err != nil {
            return fmt.Errorf("failed to parse premake5Catch2.tmpl: %v", err)
        }
        err = tmpl.Execute(testContent, data)
        if err != nil {
            return fmt.Errorf("failed to execute premake5Catch2.tmpl: %v", err)
        }
        content.WriteString(testContent.String())
	default:
	}

	if err := os.WriteFile(filepath.Join(basePath, "premake5.lua"), []byte(content.String()), 0644); err != nil {
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
	}
	if err != nil {
		return fmt.Errorf("failed to create build solution: %v", err)
	}

	return nil
}

func createProjectStructure(m model) error {
	basePath := filepath.Join(".", m.ProjectName)
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

func generateGTestFiles(m model) error {
	basePath := filepath.Join(".", m.ProjectName)
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

func generateCatch2Files(m model) error {
    basePath := filepath.Join(".", m.ProjectName)
    content := `#include <catch2/catch_test_macros.hpp>
#include "foo/foo.hpp"

// Test case for the bar function
TEST_CASE("SumTest HandlesPositiveInput", "[SumTest]") {
    REQUIRE(foo::bar(40, 2) == 42);
}

TEST_CASE("SumTest HandlesNegativeInput", "[SumTest]") {
    REQUIRE(foo::bar(-3, 2) == -1);
}

TEST_CASE("SumTest HandlesZeroInput", "[SumTest]") {
    REQUIRE(foo::bar(0, 0) == 0);
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
func initNewProject(ProjectName string) {
	m := initialModel()
	p := tea.NewProgram(m)
	output, err := p.Run()
	if err != nil {
		fmt.Println(err)
	}
	m = output.(model)
	m.ProjectName = ProjectName
	err = createProjectStructure(m)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Project created successfully")
}

func showUsage() {
	fmt.Println("Usage: gengar [init] <project-name>")
}

var argIndex = 1
func GetArg(L *lua.LState) int {
    argIndex++
    if len(os.Args) > argIndex {
        L.Push(lua.LString(os.Args[argIndex]))
        return 1
    }
    return 0
}


func main() {
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		templatePath = filepath.Join("/usr/local/share/gengar", "templates")
	}

    if _, err := os.Stat(scriptsPath); os.IsNotExist(err) {
        scriptsPath = filepath.Join("/usr/local/share/gengar", "scripts")
    }

	ProjectName := ""
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
			ProjectName = os.Args[2]
		} else {
			fmt.Println("Please provide a project name")
			os.Exit(1)
		}

		initNewProject(ProjectName)

	case "help":
		showUsage()
	default:
		// Search for a lua file on the scrips directory
        // If the file is found, execute it 

        L := lua.NewState()
        defer L.Close()
        L.SetGlobal("getArg", L.NewFunction(GetArg))
        if err := L.DoFile(filepath.Join(scriptsPath, commandOption + ".lua")); err != nil {
            if err.Error() == "open scripts/" + commandOption + ".lua: no such file or directory" {    
                fmt.Println("Command not found")
                showUsage()
            } else {
                fmt.Println("Error executing script: ", err)
            }
        }

	}

}
