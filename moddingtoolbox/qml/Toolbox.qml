import QtQuick 2.4
import QtQuick.Controls 1.3

ApplicationWindow {
	title: "Modding Toolbox"
	width: 340
	height: 240

	TabView {
		anchors.fill: parent

		Tab {
			title: "Hash"
			anchors.margins: 6

			Hasher {}
		}

		Tab {
			title: "Convert"
			anchors.margins: 6

			Converter {}
		}
	}
}
