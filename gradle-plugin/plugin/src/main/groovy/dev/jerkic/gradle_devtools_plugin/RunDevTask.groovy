package dev.jerkic.gradle_devtools_plugin

import org.gradle.api.DefaultTask
import org.gradle.api.tasks.TaskAction

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
