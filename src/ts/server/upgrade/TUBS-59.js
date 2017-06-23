db.User.update({data: {$exists: true}}, {$rename: {"data": "Avatar"}, $unset: {"kind": true}}, {multi: true});
