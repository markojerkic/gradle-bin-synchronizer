# Support for SpringBoot Devtools with Gradle and Jdtls IDEs

SpringBoot Devtools does not play nice with Jdtls IDEs like Eclipse and VSCode and Neovim (btw).
This is because Devtools uses a custom classloader to reload classes, which is not supported by Jdtls.
This project provides a workaround by using a separate process to watch for changes in the source directory and sync them to the target directory.

> Caution: This is a workaround and not a perfect solution. It makes a lot of assumptions and may not work in all cases.

## Recomened gradle script

```groovy
import groovy.json.JsonSlurper

class BinaryManager {
    final project
    final repoOwner = "markojerkic"
    final repoName = "gradle-bin-synchronizer"
    final baseDir = ".gradle/devtools-watcher"

    BinaryManager(project) {
        this.project = project
    }

    String getBinaryPath() {
        def osName = System.getProperty("os.name").toLowerCase()
        def binaryName = osName.contains("windows") ?
            "devtools-watcher-windows-amd64.exe" :
            "devtools-watcher-linux-amd64"

        return "${baseDir}/${binaryName}"
    }

    String getLatestVersion() {
        def apiUrl = "https://api.github.com/repos/${repoOwner}/${repoName}/releases/latest"
        def connection = new URL(apiUrl).openConnection() as HttpURLConnection
        connection.requestMethod = 'GET'
        connection.setRequestProperty("Accept", "application/vnd.github.v3+json")

        if (connection.responseCode == 200) {
            def response = new JsonSlurper().parse(connection.inputStream)
            return response.tag_name
        }
        return null
    }

    void downloadIfNeeded() {
        def binary = new File(project.rootDir, binaryPath)
        def versionFile = new File(project.rootDir, "${baseDir}/version")
        def latestVersion = getLatestVersion()

        if (!binary.exists() || !versionFile.exists() || versionFile.text != latestVersion) {
            project.logger.lifecycle("Downloading devtools watcher version ${latestVersion}...")
            downloadBinary(latestVersion)
            versionFile.parentFile.mkdirs()
            versionFile.text = latestVersion
            binary.setExecutable(true)
        }
    }

    private void downloadBinary(String version) {
        def binaryName = getBinaryPath().split('/')[-1]
        def downloadUrl = "https://github.com/${repoOwner}/${repoName}/releases/download/${version}/${binaryName}"

        def binary = new File(project.rootDir, binaryPath)
        // Delete the binary if it exists
        binary.delete()
        binary.parentFile.mkdirs()
        project.logger.lifecycle("Downloading ${downloadUrl} to ${binary.absolutePath}...")

        new URL(downloadUrl).openStream().withStream { input ->
            binary.withOutputStream { output ->
                output << input
            }
        }
        project.logger.lifecycle("Downloaded ${binary.absolutePath}")
    }

    Process startWatcher(String sourceDir, String targetDir) {
        def command = [
            new File(project.rootDir, binaryPath).absolutePath,
            "-watchDir", sourceDir,
            "-syncDir", targetDir
        ]

        return new ProcessBuilder(command)
            .redirectOutput(ProcessBuilder.Redirect.INHERIT)
            .redirectError(ProcessBuilder.Redirect.INHERIT)
            .start()
    }
}

class RunDevTask extends DefaultTask {
    private Process watcherProcess

    RunDevTask() {
        finalizedBy project.tasks.bootRun
    }

    @TaskAction
    void execute() {
        def binaryManager = new BinaryManager(project)

        // Download binary if needed (this can be done in parallel later)
        binaryManager.downloadIfNeeded()

        // Start the watcher process
        watcherProcess = binaryManager.startWatcher(
            "bin/main/java",
            "build/classes/java/main"
        )

        // Add shutdown hook to clean up the watcher process
        Runtime.runtime.addShutdownHook(new Thread({
            watcherProcess?.destroyForcibly()
        } as Runnable))
    }
}

tasks.register('runDev', RunDevTask) {
    group = 'application'
    description = 'Runs the application in development mode with auto-reload'

    dependsOn tasks.classes
    tasks.bootRun.mustRunAfter tasks.classes

    tasks.bootRun {
        standardInput = System.in
        systemProperty "spring.profiles.active", "development"
    }
}
```

Place this in `gradle/devtools-watcher.gradle` and apply it in your `build.gradle` like this:

```groovy
apply from: 'gradle/devtools-watcher.gradle'
```

Then, you can run the app with:
```shell
./gradlew runDev
```
