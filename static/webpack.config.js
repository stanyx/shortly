var path = require("path");

module.exports = {
  resolve: {
    extensions: ['.ts', '.tsx', '.js']
  },
  entry: "./src/index.tsx",
  output: {
    filename: "bundle.js",
    path: path.resolve(__dirname, "dist"),
    publicPath: "/dist/"
  },
  module: {
    rules: [
      { test: /\.tsx?$/, loader: "ts-loader" },
      { test:/\.css$/, use:['style-loader','css-loader']}
    ]
  },
  devServer: {
    stats: {
      assets: false,
      hash: false,
      chunks: false,
      errors: true,
      errorDetails: true,
    },
    overlay: true
  }
};