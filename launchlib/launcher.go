/* Copyright 2015 Palantir Technologies, Inc. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package launchlib

import (
	"fmt"
	"os"
	"path"
	"strings"
	"syscall"
)

// Returns explicitJavaHome if it is not the empty string, or the value of the JAVA_HOME environment variable otherwise.
// Panics if neither of them is set.
func getJavaHome(explicitJavaHome string) string {
	if explicitJavaHome != "" {
		return explicitJavaHome
	}

	javaHome := os.Getenv("JAVA_HOME")
	if len(javaHome) == 0 {
		panic("JAVA_HOME environment variable not set")
	}
	return javaHome
}

func getWorkingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
}

// Prepends each of the given classpath entries with the given working directory.
func absolutizeClasspathEntries(workingDir string, relativeClasspathEntries []string) []string {
	absoluteClasspathEntries := make([]string, len(relativeClasspathEntries))
	for i, entry := range relativeClasspathEntries {
		absoluteClasspathEntries[i] = path.Join(workingDir, entry)
	}
	return absoluteClasspathEntries
}

func joinClasspathEntries(classpathEntries []string) string {
	return strings.Join(classpathEntries, ":")
}

func Launch(staticConfig *StaticLauncherConfig, customConfig *CustomLauncherConfig) {
	fmt.Printf("Launching with static configuration %v and custom configuration %v\n", *staticConfig, *customConfig)

	workingDir := getWorkingDir()
	fmt.Println("Working directory:", workingDir)

	javaHome := getJavaHome(staticConfig.JavaHome)
	fmt.Println("Using JAVA_HOME:", javaHome)
	javaCommand := path.Join(javaHome, "/bin/java")

	classpath := joinClasspathEntries(absolutizeClasspathEntries(workingDir, staticConfig.Classpath))
	fmt.Println("Classpath:", classpath)

	var args []string
	args = append(args, javaCommand) // 0th argument is the command itself
	args = append(args, staticConfig.JvmOpts...)
	args = append(args, customConfig.JvmOpts...)
	args = append(args, "-classpath", classpath)
	args = append(args, staticConfig.MainClass)
	args = append(args, staticConfig.Args...)
	fmt.Printf("Argument list to Java binary: %v\n\n", args)

	execWithChecks(javaCommand, args)
}

func execWithChecks(javaExecutable string, args []string) {
	env := os.Environ()
	execErr := syscall.Exec(javaExecutable, args, env)
	if execErr != nil {
		if os.IsNotExist(execErr) {
			fmt.Println("Java Executable not found at:", javaExecutable)
		}
		panic(execErr)
	}
}
