module.exports = async function (context, req) {
    context.log('Setup function called');

    context.res = {
        status: 200,
        body: "Hello from Setup Function!"
    };
};