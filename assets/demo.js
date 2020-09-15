// speed hack
//L.CRS.scale = L.CRS.EPSG3857.scale = L.CRS.EPSG3395.scale = L.CRS.EPSG4326.scale = L.CRS.EPSG900913.scale = L.CRS.Earth.scale = function(t) { return 1 << (t+8); };
//L.CRS.Simple = function(t) { return 1 << t; };
L.CRS.scale = L.CRS.EPSG3857.scale = L.CRS.EPSG3395.scale = L.CRS.EPSG4326.scale = L.CRS.EPSG900913.scale = L.CRS.Earth.scale = function(t) { return 256 * (2 ** t); };
L.CRS.Simple = function(t) { return 2 ** t; };

function getRequestParm(name) {
	var re = location.search.match('[?&]'+encodeURIComponent(name)+'=([^&]*)');
	if(re) {
		return decodeURIComponent(re[1]);
	}
	return false;
}

var initPos = {
	lat: 23.83,
	lng: 121.2,
	zoom: 8,
}

var loc = getRequestParm('loc');
if(loc) {
	var p = loc.split(',');
	initPos.lat = p[0];
	initPos.lng = p[1];
	initPos.zoom = p[2];
}

var map = L.map('map', {
	center: [initPos.lat, initPos.lng],
	zoom: initPos.zoom,
	zoomControl: false,
	attributionControl: true,
});
L.control.zoom( {position: 'topright'} ).addTo(map);
L.control.scale().addTo(map);

map.on('moveend', function (e) {
	var lat = map.getCenter().lat.toFixed(6);
	var lng = map.getCenter().lng.toFixed(6);
	var zoom = map.getZoom();
	var parms = "loc=" + lat + "," + lng + "," + zoom;
	history.replaceState(null, document.title, "?" + parms);
});

var baseMap = {
	'臺灣通用電子地圖': L.tileLayer('https://wmts.nlsc.gov.tw/wmts/EMAP/default/GoogleMapsCompatible/{z}/{y}/{x}', {
		maxZoom: 18,
		id: 'EMAP',
	}),
	'臺灣通用電子地圖(黑底)': L.tileLayer('https://wmts.nlsc.gov.tw/wmts/EMAP2/default/GoogleMapsCompatible/{z}/{y}/{x}', {
		maxZoom: 18,
		id: 'EMAP2',
	}),
	'OSM': L.tileLayer('https://gis.sinica.edu.tw/worldmap/file-exists.php?img=OSM-png-{z}-{x}-{y}', {
		maxZoom: 18,
		id: 'OSM',
		attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
	}),
	'Carto Dark': L.tileLayer('https://cartodb-basemaps-{s}.global.ssl.fastly.net/dark_all/{z}/{x}/{y}.png', {
		maxZoom: 18,
		id: 'Carto-Dark',
		attribution: "Map tiles by Carto, under CC BY 3.0. Data by OpenStreetMap, under ODbL."
	}),
};
baseMap['臺灣通用電子地圖'].addTo(map);
var layerControl = L.control.layers(baseMap, {}, {
	collapsed: false,
})
layerControl.addTo(map);

