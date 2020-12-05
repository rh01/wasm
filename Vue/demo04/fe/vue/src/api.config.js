const isPro = Object.is(process.env.NODE_ENV, 'production')

module.exports = {
    baseUrl: isPro ? 'https://cs.aliyuncs.com' : 'https://cs.aliyuncs.com'
}