import QtQuick 2.4
import QtQuick.Controls 1.3
import QtQuick.Dialogs 1.2

ApplicationWindow {
	FileDialog {
		id: casPartDialog
		title: "Please choose a package"
		nameFilters: [ "Package files (*.package)" ]
	}

	Binding {
		target: app
		property: "casPartFile"
		value: casPartDialog.fileUrl
	}
	
	FileDialog {
		id: thumbnailDialog
		title: "Please choose a package"
		nameFilters: [ "Package files (*.package)" ]
	}

	Binding {
		target: app
		property: "thumbFile"
		value: thumbnailDialog.fileUrl
	}

	FileDialog {
		id: exportDirDialog
		title: "Please choose a directory"
		selectFolder: true
	}

	Binding {
		target: app
		property: "exportDir"
		value: exportDirDialog.fileUrl
	}

	property real windowMargin: 8
	property real windowSpacing: 4
	property real fileNameWidth: 250

	title: "Thumbnail Extractor"
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

		Label { text: "Cas Part Package:" }

		Row {
			spacing: windowSpacing

			TextField {
				text: casPartDialog.fileUrl
				width: fileNameWidth
				enabled: false
			}

			Button {
				text: "Browse"
				onClicked: { casPartDialog.open() }
			}
		}

		Label { text: "Thumbnail Package:" }

		Row {
			spacing: windowSpacing

			TextField {
				text: thumbnailDialog.fileUrl
				width: fileNameWidth
				enabled: false
			}

			Button {
				text: "Browse"
				onClicked: { thumbnailDialog.open() }
			}
		}

		Label { text: "Export Directory:" }

		Row {
			spacing: windowSpacing
			
			TextField {
				text: exportDirDialog.fileUrl
				width: fileNameWidth
				enabled: false
			}

			Button {
				text: "Browse"
				onClicked: { exportDirDialog.open() }
			}
		}

		Button {
			text: "Extract"
			anchors.right: parent.right
			onClicked: { app.export() }
		}
		
		Label { text: app.information }
	}
}
