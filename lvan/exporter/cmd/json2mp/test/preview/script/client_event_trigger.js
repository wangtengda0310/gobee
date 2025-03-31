window.gameScript = (function (exports) {
    function exec_command(cmdName, context, ...args){
        return context.executeCmd(cmdName, args);
    }

    //客户端jsCode合集
	
    return exports;
}({}))