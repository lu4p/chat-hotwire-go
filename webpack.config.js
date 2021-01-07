const CopyPlugin = require('copy-webpack-plugin');
const { CleanWebpackPlugin } = require('clean-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const CssMinimizerPlugin = require('css-minimizer-webpack-plugin');
const TerserPlugin = require('terser-webpack-plugin');

var path = require("path");

module.exports = {
    entry: path.resolve(__dirname, 'js/index.js'),
    module: {
        rules: [
            {
                test: /\.css$/i,
                exclude: /node_modules/,
                use: [
                    MiniCssExtractPlugin.loader,
                    'css-loader',
                    'postcss-loader',
                ]
            },
        ],
    },
    optimization: {
        nodeEnv: 'production', // only minify in production
        minimizer: [
            new CssMinimizerPlugin(), // minify css
            new TerserPlugin(), // minify js 
        ],
    },
    output: {
        filename: '[name].js',
        path: path.resolve(__dirname, 'dist'),
    },
    plugins: [
        new CopyPlugin({
            patterns: [
                { from: "assets", to: "assets" },
            ],
        }),
        new CleanWebpackPlugin(),
        new MiniCssExtractPlugin(),
    ],
};
