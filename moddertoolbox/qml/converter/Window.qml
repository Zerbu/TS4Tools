import QtQuick 2.4
import QtQuick.Controls 1.3

ApplicationWindow {
	property real windowMargin: 8
	property real windowSpacing: 4

	title: "Converter"
	width: body.width + 2 * windowMargin + 2
	height: body.height + 2 * windowMargin + 2
	minimumWidth: width
	minimumHeight: height
	maximumWidth: width
	maximumHeight: height

	Column {
		id: body
		spacing: windowSpacing
		anchors.top: parent.top
		anchors.left: parent.left
		anchors.margins: windowMargin

		Label { text: "Enter number to convert between\nhexadecimal and decimal:" }

		TextField {
			width: parent.width
			onTextChanged: { app.changeText(text) }
		}

		Label { text: "Result:" }

		TextField {
			text: app.result
			width: parent.width
			readOnly: true
			onTextChanged: {
				if (app.alignRight) {
					horizontalAlignment = TextInput.AlignRight
				} else {
					horizontalAlignment = TextInput.AlignLeft
				}
			}
		}
	}
}
