import QtQuick 2.4
import QtQuick.Controls 1.3

ApplicationWindow {
	title: "Tester Toolbox"
	width: 200
	height: 100

	Flow {
		Button {
			text: "Thumbnail Extractor"
			onClicked: { app.create("thumbextractor") }
		}
	}
}
