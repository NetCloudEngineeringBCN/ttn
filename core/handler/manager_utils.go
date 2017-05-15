// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"context"

	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

//eventSelect select the appropriate event for device is updated/created and check if the event is possible
// * If the event is an update and the AppEUI or DevEUI is changed, we should remove the device from the NetworkServer and re-add it later
func (h *handlerManager) eventSelect(ctx context.Context, dev *device.Device, lorawan *pb_lorawan.Device, appId string) (evt types.EventType, err error) {
	if dev != nil {
		evt = types.UpdateEvent
		if dev.AppEUI != *lorawan.AppEui || dev.DevEUI != *lorawan.DevEui {
			_, err = h.handler.ttnDeviceManager.DeleteDevice(ctx, &pb_lorawan.DeviceIdentifier{
				AppEui: &dev.AppEUI,
				DevEui: &dev.DevEUI,
			})
			if err != nil {
				return "", errors.Wrap(errors.FromGRPCError(err), "Broker did not delete device")
			}
		}
	} else {
		evt = types.CreateEvent
		existingDevices, err := h.handler.devices.ListForApp(appId, nil)
		if err != nil {
			return "", err
		}
		for _, existingDevice := range existingDevices {
			if existingDevice.AppEUI == *lorawan.AppEui && existingDevice.DevEUI == *lorawan.DevEui {
				return "", errors.NewErrAlreadyExists("Device with AppEUI and DevEUI")
			}
		}
	}
	return evt, nil
}

//updateDevBrk Update the device in the Broker (NetworkServer)
func (h *handlerManager) updateDevBrk(ctx context.Context, token string, dev *device.Device, lorawan *pb_lorawan.Device) error {
	nsUpdated := dev.GetLoRaWAN()
	nsUpdated.FCntUp = lorawan.FCntUp
	nsUpdated.FCntDown = lorawan.FCntDown
	_, err := h.handler.ttnDeviceManager.SetDevice(ttnctx.OutgoingContextWithToken(ctx, token), nsUpdated)
	return err
}