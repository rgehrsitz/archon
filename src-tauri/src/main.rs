// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use serde::Serialize;
use tauri::{CustomMenuItem, Menu, MenuItem, Submenu, Window};

#[derive(Clone, Serialize)]
struct Payload {
    message: String,
}

#[tauri::command]
fn emit_event(window: Window) {
    window
        .emit(
            "open-project-event",
            Payload {
                message: "hello".into(),
            },
        )
        .unwrap();
}

fn main() {
    // here `"quit".to_string()` defines the menu item id, and the second parameter is the menu item label.
    let new_project = CustomMenuItem::new("new_project".to_string(), "New Project");
    let open_project = CustomMenuItem::new("open_project".to_string(), "Open Project");
    let close_project = CustomMenuItem::new("close_project".to_string(), "Close Project");
    let save_project = CustomMenuItem::new("save_project".to_string(), "Save Project");
    let save_as_project = CustomMenuItem::new("save_as_project".to_string(), "Save As Project");
    let about = CustomMenuItem::new("about".to_string(), "About");

    let exit = CustomMenuItem::new("exit".to_string(), "Exit");

    let file_menu = Submenu::new(
        "File",
        Menu::new()
            .add_item(new_project)
            .add_item(open_project)
            .add_item(close_project)
            .add_native_item(MenuItem::Separator)
            .add_item(save_project)
            .add_item(save_as_project)
            .add_native_item(MenuItem::Separator)
            .add_item(exit),
    );
    let edit_menu = Submenu::new("Edit", Menu::new().add_native_item(MenuItem::Copy));
    let help_menu = Submenu::new("Help", Menu::new().add_item(about));

    let menu = Menu::new()
        .add_submenu(file_menu)
        .add_submenu(edit_menu)
        .add_submenu(help_menu);

    tauri::Builder::default()
        .menu(menu)
        .invoke_handler(tauri::generate_handler![emit_event])
        .on_menu_event(|event| match event.menu_item_id() {
            "exit" => {
                std::process::exit(0);
            }
            "open_project" => {
                event
                    .window()
                    .emit(
                        "open-project-event",
                        Payload {
                            message: "hello".into(),
                        },
                    )
                    .unwrap();
            }
            _ => {}
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
