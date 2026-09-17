import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { ConfirmComponent } from './common/confirm/confirm';
import { ToastComponent } from './common/toast/toast';

@Component({
  selector: 'app-root',
  templateUrl: './app.html',
  styleUrl: './app.css',
  imports: [RouterOutlet, ConfirmComponent, ToastComponent],
})
export class App {}
