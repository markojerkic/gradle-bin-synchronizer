package dev.jerkic.gradle_devtools_plugin

import groovy.json.JsonSlurper
import org.gradle.api.Project

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
