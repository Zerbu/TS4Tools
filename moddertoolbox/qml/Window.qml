import QtQuick 2.4
import QtQuick.Controls 1.3

ApplicationWindow {
	title: "Modder Toolbox"
	width: 200
	height: 100

	Flow {
		Button {
			text: "Hasher"
			onClicked: { app.create("hasher") }
		}

		Button {
			text: "Converter"
			onClicked: { app.create("converter") }
		}
	}
}
