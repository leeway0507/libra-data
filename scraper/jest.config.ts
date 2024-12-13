

const config = {
    verbose: true,
    forceExit: true,
    detectOpenHandles: true,
    transform: {
        // "\\.[jt]sx?$": "babel-jest",
        "\\.ts$": "ts-jest",

    },
    transformIgnorePatterns: [
        "/node_modules/" // If you have ESM dependencies
    ],
    extensionsToTreatAsEsm: [".ts", ".tsx"], // Treat TypeScript files as ESM
};

module.exports = config;
