package dev.jerkic.gradle_devtools_plugin

import org.gradle.api.Plugin
import org.gradle.api.Project

class DevToolsPlugin implements Plugin<Project> {
    @Override
    void apply(Project project) {
        // Configure Spring Boot task if the plugin is applied
        project.plugins.withId('org.springframework.boot') {
            def runDevTask = project.tasks.register('runDev', RunDevTask)

            runDevTask.configure {
                dependsOn('classes')
                project.tasks.bootRun.mustRunAfter('classes')
                finalizedBy('bootRun')
            }

            // Configure bootRun
            project.tasks.bootRun {
                standardInput = System.in
                systemProperty "spring.profiles.active", "development"
            }
        }
    }
}
