rrpBuildGoCode {
    projectKey = 'go-odata'
    testDependencies = ['mongo', 'postgres']
    buildImage = 'amr-registry.caas.intel.com/rrp/ci-go-build-image:1.12.0-alpine'
    skipBuild = true
    skipDocker = true
    protexProjectName = 'bb-go-odata'

    infra = [
        stackName: 'RSP-Codepipeline-GoOdata'
    ]

    notify = [
        slack: [ success: '#ima-build-success', failure: '#ima-build-failed' ]
    ]
}
