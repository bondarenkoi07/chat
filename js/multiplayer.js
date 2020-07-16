ws.onmessage = function (e) {
    var mode;
    var data = JSON.parse(e.data);
    for (let i = 0; i < data.length; i++){
        mode = data[i]['mode'];
        x = data[i]['x'];
        y = data[i]['y'];
        modelName = data[i]['model'];

        if (mode === 'up')
            doTerrainUp(x, y);
        else if (mode === 'down')
            doTerrainDown(x, y);
        else{
            if (modelName!=='none'){
                objectData[y][x]=modelName;
                fillModelControl(x,y);
            }
        }
    }
}
