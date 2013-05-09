<script type="text/javascript" src="https://www.google.com/jsapi"></script>
<script type="text/javascript">
    google.load("visualization", "1", {packages:["corechart"]});
    google.setOnLoadCallback(drawRespChart);
    google.setOnLoadCallback(drawAvgChart);
    google.setOnLoadCallback(drawPendingChart);
    
    if (!Array.prototype.filter){
	  Array.prototype.filter = function(fun /*, thisp*/){
	    var len = this.length;
	    if (typeof fun != "function")
	      throw new TypeError();

	    var res = new Array();
	    var thisp = arguments[1];
	    for (var i = 0; i < len; i++) {
	      if (i in this){
	        var val = this[i]; // in case fun mutates this
	        fun.call(thisp, val, i, this)
	        res.push(val);
	      }
	    }

	    return res;
	  };
	}

    var allData = eval({{.Data}})
    var allData2 = eval({{.Data}})
    var allData3 = eval({{.Data}})

    function countFilter(element, index, array){
    	return element.splice(5, 2)
    }

    function avgFilter(element, index, array){
    	element[6] /= 1000 
    	return element.splice(1, 5)
    }

    function pendingFilter(element, index, array){
    	element.splice(1,4)
	return element.splice(2,1)
    }

    function drawRespChart() {
    	var arr = allData.filter(countFilter)
	    var data = google.visualization.arrayToDataTable(arr);

	    var options = {
	        title: {{.Title}}
	    };

	    var chart = new google.visualization.LineChart(document.getElementById('resp_div'));
	    chart.draw(data, options);
    }

    function drawPendingChart() {
    	var arr = allData3.filter(pendingFilter)
	    var data = google.visualization.arrayToDataTable(arr);

	    var options = {
	        title: "Pending Request"
	    };

	    var chart = new google.visualization.LineChart(document.getElementById('pending_div'));
	    chart.draw(data, options);
    }

    function drawAvgChart() {
    	var arr = allData2.filter(avgFilter)
	    var data = google.visualization.arrayToDataTable(arr);

	    var options = {
	        title: "Average Response Time [ms]"
	    };

	    var chart = new google.visualization.LineChart(document.getElementById('avg_div'));
	    chart.draw(data, options);
    }
 	setTimeout(function(){
 	  	window.location.reload(1);
	}, 60000);
</script>
<div id="resp_div" style="width: 700px; height: 350px;float:left;"></div>
<div id="pending_div" style="width: 700px; height: 350px;float:left;"></div>
<div id="avg_div" style="width: 700px; height: 350px;float:left;"></div>
