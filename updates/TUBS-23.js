var total = db.Device.count();
var progress = 0;

if (db.DeviceHistory.exists()) {
    db.DeviceHistory.renameCollection("DeviceEvent");
    db.DeviceEvent.createIndex({"Action": 1, "DeviceId": 1, "AddedDate": -1}, {background: true});
    db.DeviceEvent.createIndex({"AddedDate": 1}, {background: true, expireAfterSeconds: 2592000});
    db.DeviceEvent.update({}, {$set: {Action: 0}}, false, true);
}

db.Device.find().forEach(function (d) {
var isModify = false;
    if (d.History) {
        d.History = [];

        for (var i = 0; i < d.History.length; i++) {
            var h = d.History[i];
            h.DeviceId = d._id;
            h.Action = 0;

            if (h.Details.IndexOf("user") !== -1 && h.Details.IndexOf("updated") !== -1) {
                h.Action = 3;
            } else if (h.Details.indexOf("reinstall") !== -1) {
                h.Action = 2;
            } else if (h.Details.indexOf("first") !== -1) {
                h.Action = 1;
            }

            db.DeviceEvent.save(h);
        }
        d.History = null;
        delete d.History;
        isModify = true;
    }

    if (d.StartCount) {
        d.LaunchCount = d.StartCount;
        d.StartCount = null;
        delete  d.StartCount;
        isModify = true;
    }

    if (isModify) {
        db.Device.save(d);
        print(d._id);
    }

    progress++;
    if (progress % 1000 == 0) {
        print(total - progress);
    }
});

db.DeviceEvent.update({Action: {$ne: 2}, Details: /reinstall/}, {$set: {Action: 2}}, false, true);
db.DeviceEvent.update({Action: {$ne: 3}, Details: /updated/}, {$set: {Action: 3}}, false, true);
db.DeviceEvent.update({Action: {$ne: 1}, Details: /first/}, {$set: {Action: 1}}, false, true);
db.DeviceEvent.update({Action: {$ne: 0}, Details: 'launch'}, {$set: {Action: 0}}, false, true);

db.User.update({IsPremium: {$exists: true}}, {$unset: {IsPremium: ''}}, false, true);
db.User.update({PremiumExpires: {$exists: true}}, {$unset: {PremiumExpires: ''}}, false, true);

print("done");
