import QtQuick 2.4
import QtQuick.Controls 1.3

ApplicationWindow {
	title: "Hasher"
	width: 340
	height: 240
	maximumWidth: width
	maximumHeight: height
	minimumWidth: width
	minimumHeight: height

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
