
# Copilot Instructions for Existing App Development
1. App to be coded in golang gin framework.

# Instructions for Creating a New App
1. Create a new directory for your app in the ./app folder.
2. Inside your app directory, create a main.go file.
3. Implement your app logic in main.go.
4. Update the Makefile to include your new app in the build process.
    - make build
    - make run-server
    - make run-client

# App features
- Develop an application that can fetch build logs from the Jenkins server https://ci.jenkins.io
- Implement functionality to parse and analyze the build logs to identify common issues and errors.
- Provide a user-friendly interface to display the results of the analysis.
- Ensure the application can handle various types of build logs and formats.
- Implement caching to improve performance and reduce load on the Jenkins server.