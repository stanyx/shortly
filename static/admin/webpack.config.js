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
      { test:/\.css$/, use:['style-loader','css-loader']},
      {
        test: /\.(jpg|png|svg|gif|eot|woff[\d]?|ttf)$/,
        use: [
          {
            loader: 'file-loader',
            options: {
              name: '[name].[ext]',
              publicPath: 'dist/assets',
              outputPath: 'assets'
            }
          }
        ]
      }
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