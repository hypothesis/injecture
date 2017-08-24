#!groovy

@Library('pipeline-library') _

def img

node {
    stage('build') {
        checkout(scm)
        img = buildApp(name: 'hypothesis/injecture')
    }

    // TODO: Replace with onlyOnMaster
    if (env.BRANCH_NAME == 'develop') {
        stage('release') {
            releaseApp(image: img)
        }
    }
}

// TODO: Replace with onlyOnMaster
if (env.BRANCH_NAME == 'develop') {
    milestone()
    stage('qa deploy') {
        deployApp(image: img, app: 'injecture', env: 'qa')
    }

    milestone()
    stage('prod deploy') {
        input(message: "Deploy to prod?")
        milestone()
        deployApp(image: img, app: 'injecture', env: 'prod')
    }
}
